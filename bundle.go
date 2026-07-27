package tau

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm"
)

// bundleMagic marks a '.tauc' file that carries its dependencies with it. A
// plain bytecode file doesn't have it, so both kinds can be told apart and
// keep working.
var bundleMagic = []byte("TAUB\x01")

// A bundle is a compiled program together with everything it needs at run
// time: the source of every module it imports, and the shared objects those
// modules load. Running one touches nothing else on the filesystem.
type bundle struct {
	Bytecode []byte
	Modules  map[string]string
	Plugins  map[string][]byte
}

// importRe and pluginRe find the dependencies of a module.
//
// ponytail: a regexp over the source rather than a walk over the AST, because
// import and plugin only ever take a literal: a computed path can't be
// bundled anyway, whichever way we look for it.
var (
	importRe = regexp.MustCompile(`import\(\s*"([^"]+)"\s*\)`)
	pluginRe = regexp.MustCompile(`plugin\(\s*"([^"]+)"\s*\)`)
)

// Bundle compiles path and packs it with its dependencies.
func Bundle(path string) ([]byte, error) {
	bc, err := compile(path)
	if err != nil {
		return nil, err
	}

	b := bundle{
		Bytecode: bc.Encode(),
		Modules:  map[string]string{},
		Plugins:  map[string][]byte{},
	}

	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := b.collect(path, string(src)); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.Write(bundleMagic)
	if err := gob.NewEncoder(&buf).Encode(b); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// collect walks the imports of src, which lives at file, and stores the
// source of every module it reaches along with the plugins they open.
func (b bundle) collect(file, src string) error {
	for _, m := range pluginRe.FindAllStringSubmatch(src, -1) {
		if err := b.addPlugin(file, m[1]); err != nil {
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

		b.Modules[name] = string(mod)
		if err := b.collect(p, string(mod)); err != nil {
			return err
		}
	}

	return nil
}

// addPlugin stores the shared object a module opens, looked up the same way
// the runtime looks it up.
func (b bundle) addPlugin(file, name string) error {
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

// IsBundle reports whether b holds a bundle rather than plain bytecode.
func IsBundle(b []byte) bool {
	return bytes.HasPrefix(b, bundleMagic)
}

// openBundle unpacks a bundle: its modules go to the importer, its plugins to
// a directory the loader is pointed at, and the bytecode comes back ready to
// run. The returned function removes what was written.
func openBundle(raw []byte) (compiler.Bytecode, func(), error) {
	var b bundle

	if err := gob.NewDecoder(bytes.NewReader(raw[len(bundleMagic):])).Decode(&b); err != nil {
		return compiler.Bytecode{}, nil, errors.New("not a valid tau bundle")
	}
	vm.SetBundledModules(b.Modules)

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
