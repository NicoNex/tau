package mod

import (
	"bufio"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// The sum file is what stands between "a package manager" and "curl into a
// shell". It is trust on first use: the first build writes down what a version
// hashed to, every build after that checks it, and a version that changes
// under a tag stops the build instead of running.
//
// It is deliberately not a transparency log. That is infrastructure with an
// organisation behind it, and a file in the repository gets most of the
// property for none of it.

// HashDir is the hash of a module tree: every file under dir, by relative
// path, hashed and listed in order, and the list hashed in turn. Directories
// and modification times take no part, so the same content hashes the same
// however it arrived.
func HashDir(dir string) (string, error) {
	var lines []string

	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Symlinks are not followed and not hashed: what they point at may be
		// outside the tree, and a module is the files it brought with it.
		if !d.Type().IsRegular() {
			return nil
		}

		f, err := os.Open(p)
		if err != nil {
			return err
		}
		defer f.Close()

		h := sha256.New()
		if _, err := io.Copy(h, f); err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		lines = append(lines, fmt.Sprintf("%x  %s\n", h.Sum(nil), filepath.ToSlash(rel)))
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(lines)
	h := sha256.New()
	for _, l := range lines {
		io.WriteString(h, l)
	}
	return "h1:" + base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// A Sums is the contents of tau.sum, one hash per module version.
type Sums map[string]string

func sumKey(path, version string) string { return path + "@" + version }

// ReadSums reads the sum file in dir. A missing one is an empty set, not an
// error: the first build of a new project has nothing to check yet.
func ReadSums(dir string) (Sums, error) {
	f, err := os.Open(filepath.Join(dir, SumName))
	if err != nil {
		if os.IsNotExist(err) {
			return Sums{}, nil
		}
		return nil, err
	}
	defer f.Close()

	sums := Sums{}
	sc := bufio.NewScanner(f)
	for lineno := 1; sc.Scan(); lineno++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 {
			return nil, fmt.Errorf("%s line %d: expected \"path version hash\"", SumName, lineno)
		}
		sums[sumKey(fields[0], fields[1])] = fields[2]
	}
	return sums, sc.Err()
}

// Write saves the sums in dir, in order.
func (s Sums) Write(dir string) error {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	for _, k := range keys {
		path, version, _ := strings.Cut(k, "@")
		fmt.Fprintf(&b, "%s %s %s\n", path, version, s[k])
	}
	// Written beside itself and renamed over: two builds running at once
	// otherwise leave a file that is half of each, and a sum file that cannot
	// be read stops every build after them.
	final := filepath.Join(dir, SumName)
	tmp, err := os.CreateTemp(dir, "."+SumName+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(b.String()); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), final)
}

// Verify checks a fetched tree against what is written down, and writes it
// down when it is the first time this version is seen. The returned bool says
// whether the set changed and is worth saving.
func (s Sums) Verify(path, version, dir string) (changed bool, err error) {
	got, err := HashDir(dir)
	if err != nil {
		return false, err
	}

	key := sumKey(path, version)
	want, seen := s[key]
	if !seen {
		s[key] = got
		return true, nil
	}
	if want != got {
		return false, fmt.Errorf(
			"%s@%s does not hash to what %s says:\n\twant %s\n\tgot  %s\n"+
				"the tag was moved, or the copy in the cache was edited",
			path, version, SumName, want, got)
	}
	return false, nil
}
