package mod

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// A module is a directory, and the files in it are one scope: a name defined
// in any of them is a name all of them see, and the capitalised ones are what
// the module hands out. Splitting a file in two is then a matter of layout and
// nothing else, which is the whole point - as things were, sharing a helper
// between two files meant exporting it, and the shape of the source decided
// what the module made public.

// IsDirModule reports whether dir is a directory holding tau files, and so a
// module of its own.
func IsDirModule(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	files, err := Files(dir)
	return err == nil && len(files) > 0
}

// Files are the source files of the module in dir, in the order they will run.
//
// Sorted by name, because an order there has to be and that is the only one a
// reader can predict. Functions see each other whatever the order - a name
// used before its definition reserves its global and the definition fills it -
// so this decides only when top level code runs, not what is visible.
//
// Test files are left out: they are about the module, not part of it.
func Files(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || filepath.Ext(name) != ".tau" || strings.HasSuffix(name, "_test.tau") {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}

	sort.Strings(files)
	return files, nil
}
