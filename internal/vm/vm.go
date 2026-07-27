package vm

/*
#cgo CFLAGS: -g -Ofast -fopenmp -I../obj/libffi/include
#cgo LDFLAGS: -fopenmp -L../obj/libffi/lib -lm
#include <stdlib.h>
#include <stdio.h>
#include "vm.h"
#include "../obj/object.h"
#include "../compiler/bytecode.h"

static inline struct object get_global(struct pool *globals, size_t idx) {
	return globals->list[idx];
}

static inline void set_const(struct object *list, size_t idx, struct object o) {
	list[idx] = o;
}

extern char *tau_module_dir;

static inline void set_module_dir(char *dir) {
	free(tau_module_dir);
	tau_module_dir = dir;
}
*/
import "C"
import (
	"fmt"
	"os"
	"path/filepath"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
	"golang.org/x/term"
)

type VM struct {
	vm *C.struct_vm
}

type (
	State    C.struct_state
	Bookmark C.struct_bookmark
)

var (
	Consts    []obj.Object
	importTab = make(map[string]C.struct_object)
	TermState *term.State
)

func NewState() State {
	return State(C.new_state())
}

func (s State) Free() {
	C.state_dispose(C.struct_state(s))
}

func (s *State) SetConsts(consts []obj.Object) {
	s.consts.list = (*C.struct_object)(unsafe.Pointer(&consts[0]))
	s.consts.len = C.size_t(len(consts))
	s.consts.cap = C.size_t(len(consts))
}

func (s State) NumDefs() int {
	return int(s.ndefs)
}

func New(file string, bc compiler.Bytecode) VM {
	Consts = bc.Consts()
	setModuleDir(file)
	return VM{vm: C.new_vm(C.CString(file), cbytecode(bc))}
}

func NewWithState(file string, bc compiler.Bytecode, state State) VM {
	Consts = bc.Consts()
	if len(Consts) > 0 {
		state.SetConsts(Consts)
	}
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
		return []string{taupath + ".tau", taupath + ".tauc"}
	}

	exts := []string{".tau", ".tauc"}
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

// SetBundledModules hands the runtime the modules carried inside a bundle.
// They are found by the name they are imported with, before the filesystem is
// looked at, so a bundled program runs with nothing installed.
//
// ponytail: keyed by the import string, so two different modules imported
// under the same name would collide. Key by importer too if that ever bites.
func SetBundledModules(mods map[string]string) { bundled = mods }

var bundled map[string]string

func lookup(vmfile, taupath string) (string, error) {
	// The directory of the importing file, so that a module finds the ones
	// that sit next to it whatever the working directory is.
	for _, p := range lookupPaths(filepath.Dir(vmfile), taupath) {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no module named %q", taupath)
}

// setModuleDir points the plugin loader at the directory of the file about to
// run, so that a module opening a shared object next to itself finds it.
func setModuleDir(file string) {
	C.set_module_dir(C.CString(filepath.Dir(file)))
}

//export vm_exec_load_module
func vm_exec_load_module(vm *C.struct_vm, cpath *C.char) int {
	path := C.GoString(cpath)

	if path == "" {
		C.go_vm_errorf(vm, C.CString("import: no file provided"))
		return 1
	}

	// A bundled module comes with the program, so it is looked for before the
	// filesystem: that is what makes a built program run on its own.
	src, isBundled := bundled[path]

	p := path
	if !isBundled {
		var err error

		if p, err = lookup(C.GoString(vm.file), path); err != nil {
			msg := fmt.Sprintf("import: %v", err)
			C.go_vm_errorf(vm, C.CString(msg))
			return 1
		}
	}

	// Already imported: push the module and carry on, a non zero result would
	// stop the VM.
	if mod, ok := importTab[p]; ok {
		vm.stack[vm.sp] = mod
		vm.sp++
		return 0
	}

	if !isBundled {
		b, err := os.ReadFile(p)
		if err != nil {
			msg := fmt.Sprintf("import: %v", err)
			C.go_vm_errorf(vm, C.CString(msg))
			return 1
		}
		src = string(b)
	}
	b := []byte(src)

	tree, errs := parser.Parse(p, string(b))
	if len(errs) > 0 {
		m := fmt.Sprintf("import: %v", errs[0])
		C.go_vm_errorf(vm, C.CString(m))
		return 1
	}

	c := compiler.NewImport(int(vm.state.ndefs), &Consts)
	c.SetFileInfo(p, string(b))
	if err := c.Compile(tree); err != nil {
		C.go_vm_errorf(vm, C.CString(err.Error()))
		return 1
	}

	bc := c.Bytecode()
	(*State)(&vm.state).SetConsts(Consts)
	vm.state.ndefs = C.uint32_t(bc.NDefs())
	// The resolved path, so that the modules imported by this one are looked
	// up next to it, and its own directory for the plugins it opens.
	setModuleDir(p)
	defer setModuleDir(C.GoString(vm.file))
	tvm := C.new_vm_with_state(C.CString(p), cbytecode(bc), vm.state)
	defer C.vm_dispose(tvm)
	if i := C.vm_run(tvm); i != 0 {
		C.go_vm_errorf(vm, C.CString("import error"))
		return 1
	}
	vm.state = tvm.state

	mod := C.new_object()
	for name, sym := range c.Store {
		if sym.Scope == compiler.GlobalScope {
			o := C.get_global(vm.state.globals, C.size_t(sym.Index))

			if isExported(name) {
				if o._type == C.obj_object {
					C.object_set(mod, C.CString(name), C.object_to_module(o))
				} else {
					C.object_set(mod, C.CString(name), o)
				}
			}
		}
	}

	importTab[p] = mod
	vm.stack[vm.sp] = mod
	vm.sp++
	return 0
}

//export restore_term
func restore_term() {
	if TermState != nil {
		term.Restore(int(os.Stdin.Fd()), TermState)
	}
}

func init() {
	C.set_exit()
}
