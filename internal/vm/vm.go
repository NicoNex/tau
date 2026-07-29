package vm

/*
#cgo CFLAGS: -g -Ofast -I../obj/libffi/include
#cgo LDFLAGS: -L../obj/libffi/lib -lm
#include <stdlib.h>
#include <stdio.h>
#include "vm.h"
#include "../obj/object.h"
#include "../compiler/bytecode.h"
*/
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type VM struct {
	vm *C.struct_vm
}

type (
	State    C.struct_state
	Bookmark C.struct_bookmark
)

func NewState() State {
	return State(C.new_state())
}

func (s State) Free() {
	C.state_dispose(C.struct_state(s))
}

// NumConsts is how many constants the program has so far, which is where the
// ones of the next unit compiled on top of it will start.
func (s State) NumConsts() int {
	return int(s.consts.len)
}

func (s State) NumDefs() int {
	return int(s.ndefs)
}

func New(file string, bc compiler.Bytecode) VM {
	setModuleDir(file)
	return VM{vm: C.new_vm(C.CString(file), cbytecode(bc))}
}

func NewWithState(file string, bc compiler.Bytecode, state State) VM {
	return VM{vm: C.new_vm_with_state(C.CString(file), cbytecode(bc), C.struct_state(state))}
}

// Run executes the program and reports whether it ended in an error, so that
// a failing script fails the command that started it.
func (vm VM) Run() bool {
	ok := C.vm_run(vm.vm) == 0
	C.fflush(C.stdout)
	return ok
}

func (vm VM) State() State {
	return State(vm.vm.state)
}

func (vm VM) Free() {
	C.vm_dispose(vm.vm)
}

func (vm VM) LastPoppedStackObj() obj.Object {
	o := C.vm_last_popped_stack_elem(vm.vm)
	return *(*obj.Object)(unsafe.Pointer(&o))
}

func cobj(o obj.Object) C.struct_object {
	return *(*C.struct_object)(unsafe.Pointer(&o))
}

func cbytecode(bc compiler.Bytecode) C.struct_bytecode {
	return *(*C.struct_bytecode)(unsafe.Pointer(&bc))
}

func isExported(n string) bool {
	r, _ := utf8.DecodeRuneInString(n)
	return unicode.IsUpper(r)
}

// IsExported reports whether a module gives this name to whoever imports it,
// which it does when the name starts with an upper case letter.
func IsExported(n string) bool { return isExported(n) }

// SetArgs hands the command line to the program about to run, which reads it
// back as os.Args. It travels in the environment rather than as a builtin so
// that the language keeps its small set of globals: the os module is the only
// one that knows these variables exist, and it clears them once read.
//
// ponytail: one variable per argument because an argument may hold anything
// but a NUL, so no single separator would be safe to join them with.
func SetArgs(args []string) {
	os.Setenv("TAU_ARGC", strconv.Itoa(len(args)))
	for i, a := range args {
		os.Setenv(fmt.Sprintf("TAU_ARG%d", i), a)
	}
}

// searchDirs are the directories a module is looked up into, in order: next
// to the file that imports it, then the ones the stdlib is installed in. The
// TAUPATH environment variable, a list separated like PATH, comes first.
func searchDirs(vmdir string) []string {
	dirs := []string{}

	if taupath := os.Getenv("TAUPATH"); taupath != "" {
		dirs = append(dirs, filepath.SplitList(taupath)...)
	}
	dirs = append(dirs, vmdir)

	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "lib", "tau"))
	}

	return append(dirs, filepath.Join("/", "usr", "local", "lib", "tau"), filepath.Join("/", "lib", "tau"))
}

func lookupPaths(vmdir, taupath string) []string {
	taupath = filepath.Clean(taupath)

	// An absolute path is taken as it is.
	if filepath.IsAbs(taupath) {
		if filepath.Ext(taupath) != "" {
			return []string{taupath}
		}
		return []string{taupath + ".tau"}
	}

	// Only source: a '.tauc' holds bytecode, and the loader parses what it
	// reads, so offering one here could only end in a lexer error about a
	// byte nobody typed.
	exts := []string{".tau"}
	if filepath.Ext(taupath) != "" {
		exts = []string{""}
	}

	// Relative to the working directory first, that's what the user typed.
	paths := []string{}
	for _, ext := range exts {
		paths = append(paths, taupath+ext)
	}

	for _, dir := range searchDirs(vmdir) {
		for _, ext := range exts {
			paths = append(paths, filepath.Join(dir, taupath)+ext)
		}
	}

	return paths
}

// SearchDirs is searchDirs, for the tools that have to find a module the way
// the runtime finds it.
func SearchDirs(vmdir string) []string { return searchDirs(vmdir) }

// LookupModule resolves the module taupath imported by vmfile.
func LookupModule(vmfile, taupath string) (string, error) { return lookup(vmfile, taupath) }

// A BundledModule is a module compiled when the program was bundled, together
// with the globals its exported names ended up in. Both of those are what the
// compiler used to work out at import time.
type BundledModule struct {
	Bytecode []byte
	Exports  map[string]int
}

// SetBundledModules hands the runtime the modules carried inside a bundle, in
// the order they have to be loaded. A module is compiled knowing how many
// globals and constants come before it, so it only fits where it was meant to.
//
// ponytail: keyed by the import string, so two different modules imported
// under the same name would collide. Key by importer too if that ever bites.
func SetBundledModules(order []string, mods map[string]BundledModule) {
	bundledOrder, bundled = order, mods
}

var (
	bundled      map[string]BundledModule
	bundledOrder []string
)

// LoadBundled runs the modules that came with the program and puts them where
// an import will find them. They are loaded before the program rather than
// when it asks for them, because the place each one holds in the globals and
// in the constants was decided when it was compiled, and running them in
// another order would put them somewhere else.
func (vm VM) LoadBundled() error {
	for _, name := range bundledOrder {
		m, ok := bundled[name]
		if !ok {
			return fmt.Errorf("import: %q is listed in the bundle but not in it", name)
		}

		bc := compiler.DecodeBytecode(m.Bytecode)
		// The VM takes the name and frees it when it is disposed of, so it is
		// not ours to give back.
		cname := C.CString(name)

		tvm := C.new_vm_with_state(cname, cbytecode(bc), vm.vm.state)
		if i := C.vm_run(tvm); i != 0 {
			C.vm_dispose(tvm)
			return fmt.Errorf("import: %s failed to load", name)
		}
		vm.vm.state = tvm.state

		mod := C.new_object()
		for exp, idx := range m.Exports {
			o := C.get_global(vm.vm.state.globals, C.size_t(idx))
			// object_set copies the name, so this one is ours to free.
			cexp := C.CString(exp)

			if o._type == C.obj_object {
				C.object_set(mod, cexp, C.object_to_module(o))
			} else {
				C.object_set(mod, cexp, o)
			}
			C.free(unsafe.Pointer(cexp))
		}
		// modtab_put keeps a copy of the name, so this is the last use of the
		// one the VM owns.
		C.modtab_put(vm.vm.state.mods, cname, mod)
		C.vm_dispose(tvm)
	}
	return nil
}

func lookup(vmfile, taupath string) (string, error) {
	// The directory of the importing file, so that a module finds the ones
	// that sit next to it whatever the working directory is.
	for _, p := range lookupPaths(filepath.Dir(vmfile), taupath) {
		if _, err := os.Stat(p); err != nil {
			continue
		}

		// The absolute path, so that the same file reached as "m.tau" and as
		// "./dir/../m.tau" is one module and not two.
		if abs, err := filepath.Abs(p); err == nil {
			return abs, nil
		}
		return p, nil
	}
	return "", fmt.Errorf("no module named %q", taupath)
}

// setModuleDir points the plugin loader at the directory of the file about to
// run, so that a module opening a shared object next to itself finds it.
func setModuleDir(file string) {
	C.set_module_dir(C.CString(filepath.Dir(file)))
}

func init() {
	C.set_exit()
}
