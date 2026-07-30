package mod

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ponytail: a regexp and not the parser. What tidy has to find is
// import("literal"), and an import whose argument is worked out at run time is
// one no amount of parsing would resolve either - it fails with a message
// telling the author to require it by hand.
var importRe = regexp.MustCompile(`\bimport\(\s*"([^"]+)"\s*\)`)

// Imports lists the remote modules imported by the tau files under dir, in
// order and without repeats.
func Imports(dir string) ([]string, error) {
	seen := map[string]bool{}

	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Whatever was fetched is not this module's own source, and
			// neither is anything a version control system keeps.
			if name := d.Name(); name == ".git" || name == "pkg" && dir == p {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(p) != ".tau" {
			return nil
		}

		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		for _, m := range importRe.FindAllStringSubmatch(string(b), -1) {
			if IsRemote(m[1]) {
				seen[m[1]] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out, nil
}

// Used reports which of the required modules an import path in the list
// actually reaches, so that tidy can drop the ones nothing imports.
func Used(imports []string, required []string) map[string]bool {
	used := map[string]bool{}
	for _, imp := range imports {
		for _, req := range required {
			if imp == req || strings.HasPrefix(imp, req+"/") {
				used[req] = true
			}
		}
	}
	return used
}
