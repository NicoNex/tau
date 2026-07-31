package mod

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// A Resolver answers where the file behind a remote import path is, having
// worked out once which version of every module the build uses.
type Resolver struct {
	// Root is the directory holding the tau.mod of the program being run, or
	// "" when it is a loose script outside any module.
	Root string
	// Dirs maps a module path to the directory its version was fetched into.
	Dirs map[string]string
	// Paths are the keys of Dirs, kept so that an import can be split into
	// the module holding it and the rest.
	Paths []string
}

// Load works out the modules a program uses and makes sure they are on disk.
//
// The selection is minimum version selection, the rule Go settled on: every
// requirement in the graph names the lowest version its author wrote against,
// and the build takes the highest of those. It needs no solver, it cannot
// backtrack, and it gives the same answer today and in a year, which is more
// than can be said for resolving constraints.
func Load(dir string) (*Resolver, error) {
	r := &Resolver{Root: RootOf(dir), Dirs: map[string]string{}}
	if r.Root == "" {
		return r, nil
	}

	f, err := ParseFile(filepath.Join(r.Root, FileName))
	if err != nil {
		return nil, err
	}

	// Highest version anybody in the graph asks for, module by module.
	selected := map[string]string{}
	queue := append([]Requirement(nil), f.Require...)

	for len(queue) > 0 {
		req := queue[0]
		queue = queue[1:]

		if cur, ok := selected[req.Path]; ok && CompareVersions(req.Version, cur) <= 0 {
			continue
		}
		selected[req.Path] = req.Version

		// The requirements of that version, which may raise the selection of
		// something already seen and put it back through this loop.
		moddir, err := Fetch(req.Path, req.Version)
		if err != nil {
			return nil, err
		}
		sub, err := ParseFile(filepath.Join(moddir, FileName))
		if err != nil {
			if os.IsNotExist(err) {
				// A module without a manifest requires nothing, which is a
				// reasonable thing for a small one to be.
				continue
			}
			return nil, err
		}
		queue = append(queue, sub.Require...)
	}

	sums, err := ReadSums(r.Root)
	if err != nil {
		return nil, err
	}
	dirty := false

	for p, v := range selected {
		moddir, err := Fetch(p, v)
		if err != nil {
			return nil, err
		}
		changed, err := sums.Verify(p, v, moddir)
		if err != nil {
			return nil, err
		}
		dirty = dirty || changed

		r.Dirs[p] = moddir
		r.Paths = append(r.Paths, p)
	}

	if dirty {
		if err := sums.Write(r.Root); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// Resolve turns a remote import path into the file that holds it.
//
// One rule, everywhere: a path names either the file of that name, or the
// directory of that name holding a file called after it. So
// github.com/NicoNex/example is example.tau at the root of that repository,
// and .../example/util is either util.tau beside it or util/util.tau. The
// directory form is the one a module made of several files will grow into.
func (r *Resolver) Resolve(importPath string) (string, error) {
	notRequired := fmt.Errorf("%s is not required by any %s: run `tau get %s`",
		importPath, FileName, importPath)

	if r == nil || len(r.Paths) == 0 {
		return "", notRequired
	}
	modpath, sub, ok := Split(importPath, r.Paths)
	if !ok {
		return "", notRequired
	}

	moddir := r.Dirs[modpath]
	if sub == "" {
		// The module's own path: the file named after it, and otherwise the
		// root of the repository when that holds tau files.
		if file := filepath.Join(moddir, path.Base(modpath)+".tau"); fileExists(file) {
			return file, nil
		}
		if IsDirModule(moddir) {
			return moddir, nil
		}
		return "", fmt.Errorf("%s: %s holds no tau file", importPath, modpath)
	}

	dir := filepath.Join(moddir, filepath.FromSlash(sub))
	if file := dir + ".tau"; fileExists(file) {
		return file, nil
	}
	if IsDirModule(dir) {
		return dir, nil
	}
	return "", fmt.Errorf("%s: %s holds neither %s.tau nor a %s directory",
		importPath, modpath, sub, sub)
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
