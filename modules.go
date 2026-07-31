package tau

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NicoNex/tau/internal/mod"
	"github.com/NicoNex/tau/internal/vm"
)

// The commands that look after tau.mod. Fetching happens here and nowhere
// else: a program that is running must never reach for the network to find
// out what it is made of.

// ModInit writes a tau.mod for the module at the working directory.
func ModInit(path string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, mod.FileName)); err == nil {
		return fmt.Errorf("%s already exists", mod.FileName)
	}
	if path == "" {
		return fmt.Errorf("tau mod init: give the module a path, e.g. github.com/you/thing")
	}

	f := &mod.File{Module: path, Tau: tauLine()}
	if err := f.Write(dir); err != nil {
		return err
	}
	fmt.Printf("wrote %s for %s\n", mod.FileName, path)
	return nil
}

// Get adds a module to tau.mod and fetches it. The argument is an import path
// with an optional version, "github.com/you/thing@v1.2.0"; without one the
// highest released version is taken.
func Get(arg string) error {
	root, f, err := openModule()
	if err != nil {
		return err
	}

	path, version, _ := strings.Cut(arg, "@")
	if path == "" {
		return fmt.Errorf("tau get: give a module path")
	}
	if !mod.IsRemote(path) {
		return fmt.Errorf("tau get: %q is not a remote path, those start with a host like github.com", path)
	}
	if !mod.HaveGit() {
		return fmt.Errorf("tau get: git is needed to fetch a module and is not installed")
	}

	// What the author typed may reach inside the repository, and it is the
	// repository that has versions.
	repo, err := mod.RepoRoot(path)
	if err != nil {
		return err
	}

	if version == "" {
		if version, err = mod.Latest(repo); err != nil {
			return err
		}
	} else if !mod.ValidVersion(version) {
		return fmt.Errorf("tau get: %q is not a version like v1.2.3", version)
	} else if !mod.MatchesMajor(repo, version) {
		// Two majors of a library are two modules, and the path is what tells
		// them apart. Asking for v2 of a path with no suffix is asking for
		// something that path does not name.
		major, base := mod.PathMajor(repo)
		if major == 0 {
			return fmt.Errorf("tau get: %s is v0 and v1 of that module; for %s ask for %s/v%d@%s",
				repo, version, base, mod.Major(version), version)
		}
		return fmt.Errorf("tau get: %s is v%d of %s, and %s is not", repo, major, base, version)
	}

	f.SetRequire(repo, version)
	if err := f.Write(root); err != nil {
		return err
	}

	// Loading fetches everything the new requirement drags in and writes the
	// sums down, so that what `tau get` leaves behind is a working build.
	if _, err := mod.Load(root); err != nil {
		return err
	}
	fmt.Printf("%s %s\n", repo, version)
	return nil
}

// ModDownload fetches everything tau.mod requires, so that a later build needs
// no network.
func ModDownload() error {
	root, _, err := openModule()
	if err != nil {
		return err
	}
	r, err := mod.Load(root)
	if err != nil {
		return err
	}
	for _, p := range r.Paths {
		fmt.Println(p, r.Dirs[p])
	}
	return nil
}

// ModTidy makes tau.mod say what the source actually imports: what is missing
// is added at its latest version, what nothing imports any more is dropped.
func ModTidy() error {
	root, f, err := openModule()
	if err != nil {
		return err
	}

	imports, err := mod.Imports(root)
	if err != nil {
		return err
	}

	// Everything imported has to be required by something. The longest
	// requirement that covers an import wins, and an import nothing covers
	// needs its repository found.
	var required []string
	for _, r := range f.Require {
		required = append(required, r.Path)
	}

	for _, imp := range imports {
		if _, _, ok := mod.Split(imp, required); ok {
			continue
		}
		if !mod.HaveGit() {
			return fmt.Errorf("tau mod tidy: %s is imported but not required, and git is not installed to look it up", imp)
		}
		repo, err := mod.RepoRoot(imp)
		if err != nil {
			return err
		}
		version, err := mod.Latest(repo)
		if err != nil {
			return err
		}
		f.SetRequire(repo, version)
		required = append(required, repo)
		fmt.Printf("+ %s %s\n", repo, version)
	}

	used := mod.Used(imports, required)
	kept := f.Require[:0]
	for _, r := range f.Require {
		if used[r.Path] {
			kept = append(kept, r)
			continue
		}
		fmt.Printf("- %s %s\n", r.Path, r.Version)
	}
	f.Require = kept

	if err := f.Write(root); err != nil {
		return err
	}
	_, err = mod.Load(root)
	return err
}

// CheckImports makes sure every module the program names can be found, before
// the program runs a single instruction.
//
// The imports are walked the way `tau build` already walks them, and for the
// same reason: an import takes a literal, so the set is knowable without
// running anything. What this adds is that `tau run` now agrees with `tau
// build` about it. A path that does not resolve was going to be an error
// either way; the difference is whether it arrives now or halfway through a
// run, on the one branch nobody tested.
//
// It is also where the modules from elsewhere are fetched from the cache and
// checked against tau.sum. Doing that here means it happens once, at the
// start, rather than the first time execution wanders into an import.
func CheckImports(entry string) error {
	entry, err := filepath.Abs(entry)
	if err != nil {
		return err
	}

	var (
		seen    = map[string]bool{}
		remotes []string
		walk    func(file string) error
	)

	// A module may be a directory of files, and every one of them imports on
	// its own account.
	var walkModule func(p string) error

	walkModule = func(p string) error {
		info, err := os.Stat(p)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return walk(p)
		}
		files, err := mod.Files(p)
		if err != nil {
			return err
		}
		for _, f := range files {
			if err := walk(f); err != nil {
				return err
			}
		}
		return nil
	}

	walk = func(file string) error {
		if seen[file] {
			return nil
		}
		seen[file] = true

		src, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		// Comments first, or the usage example at the top of a module - the
		// shape every file in the stdlib starts with - would be read as a
		// dependency, and a program would refuse to run over a line that is
		// there to be read by a person.
		for _, m := range importRe.FindAllStringSubmatch(stripComments(string(src)), -1) {
			path := m[1]

			if mod.IsRemote(path) {
				remotes = append(remotes, path)
				continue
			}
			next, err := vm.LookupModule(file, path)
			if err != nil {
				return fmt.Errorf("%v, imported by %s", err, relName(file))
			}
			if err := walkModule(next); err != nil {
				return err
			}
		}
		return nil
	}

	if err := walk(entry); err != nil {
		return err
	}
	if len(remotes) == 0 {
		return nil
	}

	// One resolver for the lot: working out which version of what the build
	// uses is the same answer for every import of one program.
	r, err := mod.Load(filepath.Dir(entry))
	if err != nil {
		return err
	}
	for i := 0; i < len(remotes); i++ {
		path := remotes[i]

		file, err := r.Resolve(path)
		if err != nil {
			return err
		}
		// A fetched module imports too, and what it imports has to be there
		// as much as anything else.
		if err := walkModule(file); err != nil {
			return err
		}
	}
	return nil
}

// relName is the shortest honest way to name a file in a message: the path the
// author would recognise when it is under the working directory, and the whole
// of it when it is somewhere else, such as the cache.
func relName(file string) string {
	wd, err := os.Getwd()
	if err != nil {
		return file
	}
	rel, err := filepath.Rel(wd, file)
	if err != nil || strings.HasPrefix(rel, "..") {
		return file
	}
	return rel
}

// tauLine is what goes on the `tau` line of a new manifest: the release this
// module is written against, without the commit that built the binary.
func tauLine() string {
	v := strings.TrimPrefix(TauVersion, "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[0] + "." + parts[1]
}

// openModule finds the tau.mod governing the working directory and reads it.
func openModule() (root string, f *mod.File, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}
	if root = mod.RootOf(dir); root == "" {
		return "", nil, fmt.Errorf("no %s here or above: run `tau mod init <path>` first", mod.FileName)
	}
	f, err = mod.ParseFile(filepath.Join(root, mod.FileName))
	return root, f, err
}
