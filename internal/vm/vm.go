package vm

// #cgo CFLAGS: -Werror -g -Ofast -mtune=native -fopenmp
// #cgo LDFLAGS: -fopenmp -lgc
// #include <stdlib.h>
// #include <stdio.h>
// #include "vm.h"
// #include "../obj/object.h"
import "C"
import (
	"fmt"
	"os"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
)

type (
	VM       = *C.struct_vm
	State    = C.struct_state
	Bookmark = C.struct_bookmark
)

var (
	Consts    []obj.Object
	importTab = make(map[string]C.struct_object)
)

func NewState() State {
	return C.new_state()
}

func New(file string, bc compiler.Bytecode) VM {
	return C.new_vm(C.CString(file), cbytecode(bc))
}

func NewWithState(file string, bc compiler.Bytecode, state State) VM {
	if len(Consts) > 0 {
		state.consts = (*C.struct_object)(unsafe.Pointer(&Consts[0]))
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

func cbytecode(bc compiler.Bytecode) C.struct_bytecode {
	return *(*C.struct_bytecode)(unsafe.Pointer(&bc))
}

func isExported(n string) bool {
	r, _ := utf8.DecodeRuneInString(n)
	return unicode.IsUpper(r)
}

//export vm_exec_load_module
func vm_exec_load_module(vm *C.struct_vm, cpath *C.char) {
	path := C.GoString(cpath)

	p, err := obj.ImportLookup(path)
	if err != nil {
		msg := fmt.Sprintf("import: %v", err)
		C.go_vm_errorf(vm, C.CString(msg))
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
	}

	tree, errs := parser.Parse(path, string(b))
	if len(errs) > 0 {
		m := fmt.Sprintf("import: multiple errors in module %s", path)
		C.go_vm_errorf(vm, C.CString(m))
	}

	c := compiler.NewImport(int(vm.state.ndefs), &Consts)
	c.SetFileInfo(path, string(b))
	if err := c.Compile(tree); err != nil {
		C.go_vm_errorf(vm, C.CString(err.Error()))
	}

	bc := c.Bytecode()
	vm.state.consts = (*C.struct_object)(unsafe.Pointer(bc.Consts()))
	vm.state.nconsts = C.uint32_t(bc.NConsts())
	vm.state.ndefs = C.uint32_t(bc.NDefs())
	tvm := C.new_vm_with_state(C.CString(path), cbytecode(bc), vm.state)
	if i := C.vm_run(tvm); i != 0 {
		C.go_vm_errorf(vm, C.CString("import error"))
	}
	vm.state = tvm.state

	mod := C.new_object()
	for name, sym := range c.Store {
		if sym.Scope == compiler.GlobalScope {
			o := vm.state.globals[sym.Index]

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

func init() {
	C.gc_init()
}
