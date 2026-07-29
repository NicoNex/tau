// Package bundle reads a program that travels with everything it needs. The
// writing side lives with the compiler, because making one needs a parser; the
// reading side is here so that a runtime built to run a bundled program can
// have neither.
package bundle

import (
	"bytes"
	"encoding/gob"
	"errors"
	"os"
	"path/filepath"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm"
)

// Magic marks a bundle. Plain bytecode does not have it, so both kinds can be
// told apart and keep working.
var Magic = []byte("TAUB\x02")

// A Bundle is a compiled program together with everything it needs at run
// time: every module it imports, already compiled, and the shared objects
// those modules load. Running one touches nothing else on the filesystem and
// asks nothing of the parser or the compiler.
type Bundle struct {
	Bytecode []byte
	// The modules in the order they have to be loaded, dependencies first. A
	// module is compiled knowing how many globals and constants come before
	// it, so its indices are absolute and only hold if it is loaded in the
	// place it was compiled for.
	Order   []string
	Modules map[string]ModuleCode
	Plugins map[string][]byte
}

// A module as it travels: its bytecode, and where in the globals its exported
// names ended up. The names are what the symbol table used to answer at run
// time, which is the last thing an import needed the compiler for.
type ModuleCode struct {
	Bytecode []byte
	Exports  map[string]int
}

// bundledModules hands the runtime what it needs of each module: the bundle
// keeps its own type so that its shape can change without the runtime caring.
func bundledModules(mods map[string]ModuleCode) map[string]vm.BundledModule {
	out := make(map[string]vm.BundledModule, len(mods))

	for name, m := range mods {
		out[name] = vm.BundledModule{Bytecode: m.Bytecode, Exports: m.Exports}
	}
	return out
}

// Run runs a compiled program, named after where it came from so that a
// runtime error can say where it happened. The modules that came with it, if
// any, are loaded into the same state first.
func Run(name string, bytecode compiler.Bytecode) error {
	tvm := vm.New(name, bytecode)

	if err := tvm.LoadBundled(); err != nil {
		return err
	}
	if !tvm.Run() {
		return errors.New("runtime error")
	}
	return nil
}

// Is reports whether b holds a bundle rather than plain bytecode.
func Is(b []byte) bool {
	return bytes.HasPrefix(b, Magic)
}

// openBundle unpacks a bundle: its modules go to the importer, its plugins to
// a directory the loader is pointed at, and the bytecode comes back ready to
// run. The returned function removes what was written.
func Open(raw []byte) (compiler.Bytecode, func(), error) {
	var b Bundle

	if err := gob.NewDecoder(bytes.NewReader(raw[len(Magic):])).Decode(&b); err != nil {
		return compiler.Bytecode{}, nil, errors.New("not a valid tau bundle")
	}
	vm.SetBundledModules(b.Order, bundledModules(b.Modules))

	clean := func() {}
	if len(b.Plugins) > 0 {
		dir, err := os.MkdirTemp("", "tau-plugins")
		if err != nil {
			return compiler.Bytecode{}, nil, err
		}
		clean = func() { os.RemoveAll(dir) }

		for name, so := range b.Plugins {
			dst := filepath.Join(dir, name)
			if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
				clean()
				return compiler.Bytecode{}, nil, err
			}
			// Executable: it is about to be dlopen'd.
			if err := os.WriteFile(dst, so, 0755); err != nil {
				clean()
				return compiler.Bytecode{}, nil, err
			}
		}

		taupath := dir
		if old := os.Getenv("TAUPATH"); old != "" {
			taupath += string(filepath.ListSeparator) + old
		}
		os.Setenv("TAUPATH", taupath)
	}

	return compiler.DecodeBytecode(b.Bytecode), clean, nil
}
