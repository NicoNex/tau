package vm

// #cgo CFLAGS: -g -Ofast -fopenmp
// #cgo LDFLAGS: -fopenmp -L${SRCDIR}/bdwgc/lib ${SRCDIR}/bdwgc/lib/libgc.a
// #include <stdlib.h>
// #include <stdio.h>
// #include "vm.h"
// #include "../obj/object.h"
//
// static inline struct object get_global(struct pool *globals, size_t idx) {
// 	return globals->list[idx];
// }
//
// static inline void set_const(struct object *list, size_t idx, struct object o) {
// 	list[idx] = o;
// }
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

type (
	VM       = *C.struct_vm
	State    = C.struct_state
	Bookmark = C.struct_bookmark
)

var (
	Consts    []obj.Object
	importTab = make(map[string]C.struct_object)
	TermState *term.State
)

// NewState returns a new empty VM state.
func NewState() State {
	return C.new_state()
}

// SetConsts copies the constants in input in the VM.
func (s *State) SetConsts(consts []obj.Object) {
	s.consts.list = (*C.struct_object)(C.realloc(
		unsafe.Pointer(s.consts.list),
		C.size_t(unsafe.Sizeof(consts[0]))*C.size_t(len(consts)),
	))
	s.consts.len = C.size_t(len(consts))
	s.consts.cap = C.size_t(len(consts))

	for i, c := range consts {
		C.set_const(s.consts.list, C.size_t(i), cobj(c))
	}
}

// Free sets the state as collectable for the GC.
func (s State) Free() {
	C.state_free(s)
}

// New returns a new Tau VM instance.
func New(file string, bc compiler.Bytecode) VM {
	return C.new_vm(C.CString(file), cbytecode(bc))
}

// NewWithState returns a new Tau VM instance with the provided state.
func NewWithState(file string, bc compiler.Bytecode, state State) VM {
	if len(Consts) > 0 {
		state.SetConsts(Consts)
	}
	return C.new_vm_with_state(C.CString(file), cbytecode(bc), state)
}

// Run makes the vm run.
func (vm VM) Run() {
	C.vm_run(vm)
	C.fflush(C.stdout)
}

// Free sets the VM as collectable for the GC.
func (vm VM) Free() {
	C.vm_free(vm)
}

// FreeAll is like Free but includes the VM state as well.
func (vm VM) FreeAll() {
	vm.State().Free()
	vm.Free()
}

// State returns the VM state.
func (vm VM) State() State {
	return vm.state
}

// LastPoppedStackObj returns the object that was popped last from the VM stack.
func (vm VM) LastPoppedStackObj() obj.Object {
	o := C.vm_last_popped_stack_elem(vm)
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

func lookup(taupath string) (string, error) {
	var paths []string
	taupath = filepath.Clean(taupath)

	if ext := filepath.Ext(taupath); ext != "" {
		paths = []string{taupath, filepath.Join("/", "lib", "tau", taupath)}
	} else {
		paths = []string{
			taupath + ".tau",
			taupath + ".tauc",
			filepath.Join("/", "lib", "tau", taupath+".tau"),
			filepath.Join("/", "lib", "tau", taupath+".tauc"),
		}
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no module named %q", taupath)
}

//export vm_exec_load_module
func vm_exec_load_module(vm *C.struct_vm, cpath *C.char) {
	path := C.GoString(cpath)

	if path == "" {
		C.go_vm_errorf(vm, C.CString("import: no file provided"))
		return
	}

	// Look for the module location in the predefined ones.
	p, err := lookup(path)
	if err != nil {
		msg := fmt.Sprintf("import: %v", err)
		C.go_vm_errorf(vm, C.CString(msg))
		return
	}

	// Check whether the module has already been imported and is in the import table.
	if mod, ok := importTab[p]; ok {
		vm.stack[vm.sp] = mod
		vm.sp++
		return
	}

	// Read the module file.
	b, err := os.ReadFile(p)
	if err != nil {
		msg := fmt.Sprintf("import: %v", err)
		C.go_vm_errorf(vm, C.CString(msg))
		return
	}

	// Parse the module and get its AST.
	tree, errs := parser.Parse(path, string(b))
	if len(errs) > 0 {
		m := fmt.Sprintf("import: multiple errors in module %s", path)
		C.go_vm_errorf(vm, C.CString(m))
		return
	}

	// Compile the AST.
	c := compiler.NewImport(int(vm.state.ndefs), &Consts)
	c.SetFileInfo(path, string(b))
	if err := c.Compile(tree); err != nil {
		C.go_vm_errorf(vm, C.CString(err.Error()))
		return
	}

	// Run the bytecode of the module in a new VM instance.
	bc := c.Bytecode()
	(&vm.state).SetConsts(Consts)
	vm.state.ndefs = C.uint32_t(bc.NDefs())
	tvm := C.new_vm_with_state(C.CString(path), cbytecode(bc), vm.state)
	if i := C.vm_run(tvm); i != 0 {
		C.go_vm_errorf(vm, C.CString("import error"))
		return
	}
	vm.state = tvm.state

	// Create a new object for the module exported symbols.
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

	// Set the newly created object in the import table and in the VM stack at
	// the current stack pointer position.
	importTab[p] = mod
	vm.stack[vm.sp] = mod
	vm.sp++
}

//export restore_term
func restore_term() {
	if TermState != nil {
		term.Restore(int(os.Stdin.Fd()), TermState)
	}
}

func init() {
	C.gc_init()
	C.set_exit()
}
