package vm

// #cgo CFLAGS: -g -Ofast -mtune=native -fopenmp
// #cgo LDFLAGS: -fopenmp
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
	"errors"
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

	ErrNoFileProvided = errors.New("no file provided")
)

func NewState() State {
	return C.new_state()
}

func (s State) Free() {
	C.state_dispose(s)
}

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

func New(file string, bc compiler.Bytecode) VM {
	return C.new_vm(C.CString(file), cbytecode(bc))
}

func setConsts(state State)

func NewWithState(file string, bc compiler.Bytecode, state State) VM {
	if len(Consts) > 0 {
		state.SetConsts(Consts)
	}
	return C.new_vm_with_state(C.CString(file), cbytecode(bc), state)
}

func (vm VM) Run() {
	C.vm_run(vm)
	C.fflush(C.stdout)
}

func (vm VM) State() State {
	return vm.state
}

func (vm VM) Free() {
	C.vm_dispose(vm)
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
		if _, err := os.Stat(p); os.IsExist(err) {
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

	p, err := lookup(path)
	if err != nil {
		msg := fmt.Sprintf("import: %v", err)
		C.go_vm_errorf(vm, C.CString(msg))
		return
	}

	if mod, ok := importTab[p]; ok {
		vm.stack[vm.sp] = mod
		vm.sp++
		return
	}

	b, err := os.ReadFile(p)
	if err != nil {
		msg := fmt.Sprintf("import: %v", err)
		C.go_vm_errorf(vm, C.CString(msg))
		return
	}

	tree, errs := parser.Parse(path, string(b))
	if len(errs) > 0 {
		m := fmt.Sprintf("import: multiple errors in module %s", path)
		C.go_vm_errorf(vm, C.CString(m))
		return
	}

	c := compiler.NewImport(int(vm.state.ndefs), &Consts)
	c.SetFileInfo(path, string(b))
	if err := c.Compile(tree); err != nil {
		C.go_vm_errorf(vm, C.CString(err.Error()))
		return
	}

	bc := c.Bytecode()
	(&vm.state).SetConsts(Consts)
	vm.state.ndefs = C.uint32_t(bc.NDefs())
	tvm := C.new_vm_with_state(C.CString(path), cbytecode(bc), vm.state)
	defer C.vm_dispose(tvm)
	if i := C.vm_run(tvm); i != 0 {
		C.go_vm_errorf(vm, C.CString("import error"))
		return
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
