package mod

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// FileName is the manifest at the root of a module.
const FileName = "tau.mod"

// SumName holds the hash of every module tree the build reads, so that the
// same version fetched twice is checked to be the same bytes.
const SumName = "tau.sum"

// A File is a parsed tau.mod:
//
//	module github.com/NicoNex/example
//
//	tau 2.1
//
//	require (
//		github.com/x/y v1.4.0
//		git.sr.ht/~z/w v0.3.1
//	)
//
// A single requirement may also be written on one line, "require path v1.2.3".
// The format is read and written by hand on purpose: a manifest that needs a
// TOML parser to be understood is a manifest nobody edits.
type File struct {
	Module  string
	Tau     string
	Require []Requirement
}

// A Requirement is one line of a require block: which module, and the lowest
// version of it this one is written against.
type Requirement struct {
	Path    string
	Version string
}

// ParseFile reads the manifest at path.
func ParseFile(path string) (*File, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := Parse(string(b))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return f, nil
}

// Parse reads a manifest out of the text of one.
func Parse(src string) (*File, error) {
	f := &File{}
	sc := bufio.NewScanner(strings.NewReader(src))
	inRequire := false

	for lineno := 1; sc.Scan(); lineno++ {
		line := strings.TrimSpace(sc.Text())
		if i := strings.Index(line, "#"); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}

		if inRequire {
			if line == ")" {
				inRequire = false
				continue
			}
			r, err := parseRequirement(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineno, err)
			}
			f.Require = append(f.Require, r)
			continue
		}

		verb, rest, _ := strings.Cut(line, " ")
		rest = strings.TrimSpace(rest)

		switch verb {
		case "module":
			if rest == "" {
				return nil, fmt.Errorf("line %d: module needs a path", lineno)
			}
			f.Module = rest

		case "tau":
			f.Tau = rest

		case "require":
			if rest == "(" {
				inRequire = true
				continue
			}
			r, err := parseRequirement(rest)
			if err != nil {
				return nil, fmt.Errorf("line %d: %w", lineno, err)
			}
			f.Require = append(f.Require, r)

		default:
			return nil, fmt.Errorf("line %d: unknown verb %q", lineno, verb)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if inRequire {
		return nil, fmt.Errorf("require block is not closed")
	}
	if f.Module == "" {
		return nil, fmt.Errorf("no module line")
	}
	return f, nil
}

func parseRequirement(line string) (Requirement, error) {
	path, version, ok := strings.Cut(line, " ")
	version = strings.TrimSpace(version)

	if !ok || version == "" {
		return Requirement{}, fmt.Errorf("require %q: no version", line)
	}
	if !ValidVersion(version) {
		return Requirement{}, fmt.Errorf("require %s: %q is not a version like v1.2.3", path, version)
	}
	return Requirement{Path: path, Version: version}, nil
}

// String writes the manifest back out, requirements sorted so that two runs
// of tidy on the same set produce the same file.
func (f *File) String() string {
	var b strings.Builder

	fmt.Fprintf(&b, "module %s\n", f.Module)
	if f.Tau != "" {
		fmt.Fprintf(&b, "\ntau %s\n", f.Tau)
	}

	switch len(f.Require) {
	case 0:
	case 1:
		fmt.Fprintf(&b, "\nrequire %s %s\n", f.Require[0].Path, f.Require[0].Version)
	default:
		reqs := append([]Requirement(nil), f.Require...)
		sort.Slice(reqs, func(i, j int) bool { return reqs[i].Path < reqs[j].Path })

		b.WriteString("\nrequire (\n")
		for _, r := range reqs {
			fmt.Fprintf(&b, "\t%s %s\n", r.Path, r.Version)
		}
		b.WriteString(")\n")
	}
	return b.String()
}

// Write saves the manifest in dir.
func (f *File) Write(dir string) error {
	return os.WriteFile(filepath.Join(dir, FileName), []byte(f.String()), 0644)
}

// SetRequire adds a requirement or raises the one already there. It never
// lowers one: a build takes the highest version anybody asked for, and asking
// again for an older one is not a reason to go back.
func (f *File) SetRequire(path, version string) {
	for i, r := range f.Require {
		if r.Path == path {
			if CompareVersions(version, r.Version) > 0 {
				f.Require[i].Version = version
			}
			return
		}
	}
	f.Require = append(f.Require, Requirement{Path: path, Version: version})
}

// ValidVersion accepts the tags this scheme understands, vMAJOR.MINOR.PATCH
// and nothing else. A pre-release or a build tag is a version whose ordering
// takes a specification to explain, and ordering versions is the one thing
// this has to get right.
func ValidVersion(v string) bool {
	_, ok := parseVersion(v)
	return ok
}

func parseVersion(v string) ([3]int, bool) {
	var out [3]int

	if !strings.HasPrefix(v, "v") {
		return out, false
	}
	parts := strings.Split(v[1:], ".")
	if len(parts) != 3 {
		return out, false
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || (len(p) > 1 && p[0] == '0') {
			return out, false
		}
		out[i] = n
	}
	return out, true
}

// Major is the first number of a version, -1 when it is not one.
func Major(v string) int {
	parsed, ok := parseVersion(v)
	if !ok {
		return -1
	}
	return parsed[0]
}

// CompareVersions orders two versions, -1, 0 or 1. Anything unparseable sorts
// below everything, so a bad version can never win a selection.
func CompareVersions(a, b string) int {
	va, oka := parseVersion(a)
	vb, okb := parseVersion(b)

	switch {
	case !oka && !okb:
		return strings.Compare(a, b)
	case !oka:
		return -1
	case !okb:
		return 1
	}

	for i := 0; i < 3; i++ {
		if va[i] != vb[i] {
			if va[i] < vb[i] {
				return -1
			}
			return 1
		}
	}
	return 0
}
