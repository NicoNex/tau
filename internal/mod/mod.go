// Package mod is where a module that does not come from this machine is
// named, found and fetched.
//
// There is no registry: an import path is a place. "github.com/NicoNex/example"
// says which host holds it and what to ask that host for, which is the whole
// reason the scheme needs nothing central to work.
package mod

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsRemote reports whether an import path names a module to be fetched.
//
// The rule is the one Go uses and it needs no configuration: the first element
// of a remote path is a host, and a host has a dot in it. "strings" and
// "crypto/sha256" have none and stay what they have always been, a lookup in
// the library and then next to the importing file.
func IsRemote(path string) bool {
	if path == "" || strings.HasPrefix(path, ".") || strings.HasPrefix(path, "/") {
		return false
	}
	first, _, _ := strings.Cut(path, "/")
	return strings.Contains(first, ".")
}

// PathMajor takes the major version a path declares off the end of it.
// "github.com/x/y/v2" is major 2 of the module whose repository is
// github.com/x/y, and a path with no suffix is major 0 or 1.
//
// It is the rule Go arrived at, and it is not decoration: two majors of one
// library are two different modules, incompatible by definition, and a program
// that reaches both through its dependencies has to be able to hold both. A
// version number that is not part of the path cannot do that.
func PathMajor(path string) (major int, base string) {
	i := strings.LastIndex(path, "/v")
	if i < 0 {
		return 0, path
	}
	suffix := path[i+2:]
	if suffix == "" || suffix == "0" || suffix == "1" || strings.HasPrefix(suffix, "0") {
		return 0, path
	}
	for _, r := range suffix {
		if r < '0' || r > '9' {
			return 0, path
		}
	}

	n := 0
	for _, r := range suffix {
		n = n*10 + int(r-'0')
	}
	return n, path[:i]
}

// MatchesMajor reports whether a version belongs to the module at this path:
// v2.x.y under a path ending in /v2, and v0.x.y or v1.x.y under one with no
// suffix at all.
func MatchesMajor(path, version string) bool {
	want, _ := PathMajor(path)
	got := Major(version)

	if want == 0 {
		return got <= 1
	}
	return got == want
}

// Home is where fetched modules live: ~/.tau, or TAUHOME when it is set.
//
// It holds pkg/, one directory per module and version, and dl/, the raw
// downloads. Everything under pkg/ is read once written: a version is what it
// was the day it was fetched, so two projects can share the tree and nothing
// ever has to be reinstalled.
func Home() (string, error) {
	if h := os.Getenv("TAUHOME"); h != "" {
		return h, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find the home directory: %w", err)
	}
	return filepath.Join(home, ".tau"), nil
}

// PkgDir is where the module at this path and version is extracted.
func PkgDir(path, version string) (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "pkg", filepath.FromSlash(path)+"@"+version), nil
}

// RootOf walks up from dir looking for the tau.mod that governs it, and
// returns the directory holding it. A file outside any module gets "".
func RootOf(dir string) string {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, FileName)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// Split takes an import path apart into the module it belongs to and the
// directory inside that module, given the modules that are known.
//
// "github.com/a/b/util" with "github.com/a/b" required resolves to that module
// and "util". The longest match wins, so a module may hold another module's
// path as a prefix without stealing its imports.
func Split(path string, known []string) (mod, sub string, ok bool) {
	best := ""
	for _, k := range known {
		if path == k || strings.HasPrefix(path, k+"/") {
			if len(k) > len(best) {
				best = k
			}
		}
	}
	if best == "" {
		return "", "", false
	}
	return best, strings.TrimPrefix(strings.TrimPrefix(path, best), "/"), true
}
