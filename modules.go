package tau

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NicoNex/tau/internal/mod"
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
