package mod

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Fetching goes through git rather than over plain HTTPS, and that buys three
// things a downloader of tarballs would each have to be taught: every forge at
// once, the ssh keys and credential helpers already on this machine, and
// therefore private repositories. The price is that git has to be installed to
// *fetch*. Building from what is already in the cache needs nothing, and
// running a built program needs nothing at all.

func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	// A prompt in the middle of a build is a hang nobody can see.
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = strings.TrimSpace(string(ee.Stderr))
		}
		if stderr != "" {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), stderr)
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

// HaveGit reports whether fetching is possible at all, so that a module which
// is merely missing from the cache can say so instead of blaming git.
func HaveGit() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// Versions lists the released versions of a module, newest last.
func Versions(path string) ([]string, error) {
	out, err := git("", "ls-remote", "--tags", "--refs", "https://"+path)
	if err != nil {
		return nil, err
	}

	var vs []string
	for _, line := range strings.Split(out, "\n") {
		_, ref, ok := strings.Cut(line, "refs/tags/")
		if !ok {
			continue
		}
		if v := strings.TrimSpace(ref); ValidVersion(v) {
			vs = append(vs, v)
		}
	}
	sort.Slice(vs, func(i, j int) bool { return CompareVersions(vs[i], vs[j]) < 0 })
	return vs, nil
}

// Latest is the highest released version of a module.
func Latest(path string) (string, error) {
	vs, err := Versions(path)
	if err != nil {
		return "", err
	}
	if len(vs) == 0 {
		return "", fmt.Errorf("%s has no version tagged vX.Y.Z", path)
	}
	return vs[len(vs)-1], nil
}

// RepoRoot finds which prefix of an import path is the repository, by asking
// the host. "github.com/a/b/util" is answered by "github.com/a/b" when b is
// the repository and util a directory inside it.
//
// ponytail: longest prefix first, one network round trip each. Paths are three
// or four elements long, and this runs in `tau get` and `tau mod tidy`, never
// in a build.
func RepoRoot(path string) (string, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("%q is not an import path a host can answer", path)
	}

	for n := len(parts); n >= 2; n-- {
		candidate := strings.Join(parts[:n], "/")
		if _, err := git("", "ls-remote", "--quiet", "--exit-code", "https://"+candidate, "HEAD"); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no repository found for %q", path)
}

// Fetch puts a module version in the cache and returns where it landed. A
// version already there is left alone: the tree under pkg/ is what it was the
// day it arrived, which is what makes it safe to share between projects.
func Fetch(path, version string) (string, error) {
	dir, err := PkgDir(path, version)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}
	if !HaveGit() {
		return "", fmt.Errorf("%s@%s is not in the cache and git is not installed to fetch it", path, version)
	}

	// Next to the destination and not in /tmp: a rename across filesystems is
	// a copy that can half happen, and a half module in the cache is a module
	// that looks fetched.
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return "", err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dir), ".tmp-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmp)

	if _, err := git("", "clone", "--quiet", "--depth", "1", "--branch", version,
		"https://"+path, tmp); err != nil {
		return "", fmt.Errorf("fetching %s@%s: %w", path, version, err)
	}
	// The history is of no use to a build and is most of the bytes.
	if err := os.RemoveAll(filepath.Join(tmp, ".git")); err != nil {
		return "", err
	}

	if err := os.Rename(tmp, dir); err != nil {
		// Somebody else fetched the same version while this was running, which
		// is fine: theirs is the same bytes as ours.
		if _, serr := os.Stat(dir); serr == nil {
			return dir, nil
		}
		return "", err
	}
	return dir, nil
}
