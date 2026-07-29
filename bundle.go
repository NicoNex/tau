package tau

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	bundlepkg "github.com/NicoNex/tau/internal/bundle"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

// bundleMagic marks a '.tauc' file that carries its dependencies with it. A
// plain bytecode file doesn't have it, so both kinds can be told apart and
// keep working.
// importRe and pluginRe find the dependencies of a module.
//
// ponytail: a regexp over the source rather than a walk over the AST, because
// import and plugin only ever take a literal: a computed path can't be
// bundled anyway, whichever way we look for it.
var (
	importRe = regexp.MustCompile(`import\(\s*"([^"]+)"\s*\)`)
	pluginRe = regexp.MustCompile(`plugin\(\s*"([^"]+)"\s*\)`)
)

// Bundle compiles path and packs it with its dependencies, each of them
// compiled as well.
func Bundle(path string) ([]byte, error) {
	bc, err := compile(path)
	if err != nil {
		return nil, err
	}

	b := bundlepkg.Bundle{
		Bytecode: bc.Encode(),
		Modules:  map[string]bundlepkg.ModuleCode{},
		Plugins:  map[string][]byte{},
	}

	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// The program owns the globals and the constants that come first, and the
	// modules take what follows in the order they are collected.
	st := &bundleState{ndefs: int(bc.NDefs()), nconsts: int(bc.NConsts())}
	if err := collectInto(&b, st, path, string(src)); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write(bundlepkg.Magic)
	if err := gob.NewEncoder(&buf).Encode(b); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// bundleState carries how much of the globals and of the constants has been
// handed out, so that the next module compiled follows the ones before it.
type bundleState struct {
	ndefs   int
	nconsts int
}

// collect walks the imports of src, which lives at file, compiles every module
// it reaches and stores the plugins they open. A module goes into the bundle
// after the ones it imports, because that is the order they will be loaded in
// and the order they were compiled for.
func collectInto(b *bundlepkg.Bundle, st *bundleState, file, src string) error {
	for _, m := range pluginRe.FindAllStringSubmatch(src, -1) {
		if err := addPlugin(b, file, m[1]); err != nil {
			return err
		}
	}

	for _, m := range importRe.FindAllStringSubmatch(src, -1) {
		name := m[1]
		if _, done := b.Modules[name]; done {
			continue
		}

		p, err := vm.LookupModule(file, name)
		if err != nil {
			return fmt.Errorf("build: %v, imported by %s", err, file)
		}

		mod, err := os.ReadFile(p)
		if err != nil {
			return err
		}

		// Its own imports first: a module that runs before what it imports
		// would not find it.
		if err := collectInto(b, st, p, string(mod)); err != nil {
			return err
		}
		// The walk may have reached this one from somewhere else meanwhile.
		if _, done := b.Modules[name]; done {
			continue
		}

		code, err := compileModule(st, p, string(mod))
		if err != nil {
			return err
		}
		b.Modules[name] = code
		b.Order = append(b.Order, name)
	}

	return nil
}

// compileModule compiles one module for the place it will hold in the program,
// and notes where its exported names land.
func compileModule(st *bundleState, path, src string) (bundlepkg.ModuleCode, error) {
	tree, errs := parser.Parse(path, src)
	if len(errs) > 0 {
		return bundlepkg.ModuleCode{}, fmt.Errorf("build: %v", errs[0])
	}

	c := compiler.NewImport(st.ndefs, st.nconsts)
	c.SetFileInfo(path, src)
	if err := c.Compile(tree); err != nil {
		return bundlepkg.ModuleCode{}, err
	}

	bc := c.Bytecode()
	exports := map[string]int{}
	for name, sym := range c.Store {
		if sym.Scope == compiler.GlobalScope && vm.IsExported(name) {
			exports[name] = sym.Index
		}
	}

	st.ndefs = int(bc.NDefs())
	st.nconsts += int(bc.NConsts())
	return bundlepkg.ModuleCode{Bytecode: bc.Encode(), Exports: exports}, nil
}

// addPlugin stores the shared object a module opens, looked up the same way
// the runtime looks it up.
func addPlugin(b *bundlepkg.Bundle, file, name string) error {
	if _, done := b.Plugins[name]; done {
		return nil
	}

	for _, dir := range append([]string{filepath.Dir(file)}, vm.SearchDirs(filepath.Dir(file))...) {
		so, err := os.ReadFile(filepath.Join(dir, name))
		if err == nil {
			b.Plugins[name] = so
			return nil
		}
	}

	return fmt.Errorf("build: no plugin named %q, opened by %s", name, file)
}

