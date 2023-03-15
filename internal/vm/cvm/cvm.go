package cvm

// #cgo CFLAGS: -Werror -Wall -g -Ofast -mtune=native -fopenmp
// #cgo LDFLAGS: -fopenmp
// #include <stdlib.h>
// #include "vm.h"
// #include "obj.h"
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
	"github.com/NicoNex/tau/internal/tauerr"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type CVM = *C.struct_vm

var (
	consts    []obj.Object
	importTab = make(map[string]C.struct_object)
)

func New(file string, bc *compiler.Bytecode) CVM {
	consts = bc.Constants
	return C.new_vm(C.CString(file), cbytecode(bc))
}

func (cvm CVM) Run() {
	C.vm_run(cvm)
}

func cbytecode(bc *compiler.Bytecode) C.struct_bytecode {
	return C.struct_bytecode{
		insts:     (*C.uchar)(unsafe.Pointer(&bc.Instructions[0])),
		len:       C.uint32_t(len(bc.Instructions)),
		consts:    cObjs(bc.Constants),
		nconsts:   C.uint32_t(len(bc.Constants)),
		bookmarks: cBookmarks(bc.Bookmarks),
		bklen:     C.uint32_t(len(bc.Bookmarks)),
		ndefs:     C.uint32_t(bc.NumDefs),
	}
}

func cObjs(objects []obj.Object) *C.struct_object {
	var objs = make([]cobj.CObj, len(objects))

	for i, o := range objects {
		objs[i] = o.(cobj.CObj)
	}

	return (*C.struct_object)(unsafe.Pointer(&objs[0]))
}

func cBookmarks(bmarks []tauerr.Bookmark) *C.struct_bookmark {
	var ret = make([]C.struct_bookmark, len(bmarks))

	for i, b := range bmarks {
		ret[i] = C.struct_bookmark{
			offset: C.uint(b.Offset),
			lineno: C.uint(b.LineNo),
			pos:    C.uint(b.Pos),
			len:    C.size_t(len(b.Line)),
			line:   C.CString(b.Line),
		}
	}

	return &ret[0]
}

func isExported(n string) bool {
	r, _ := utf8.DecodeRuneInString(n)
	return unicode.IsUpper(r)
}

//export VMExecLoadModule
func VMExecLoadModule(vm *C.struct_vm, cpath *C.char) {
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

	c := compiler.NewImport(int(vm.state.ndefs), &consts)
	c.SetUseCObjects(true)
	c.SetFileInfo(path, string(b))
	if err := c.Compile(tree); err != nil {
		C.go_vm_errorf(vm, C.CString(err.Error()))
	}

	bc := cbytecode(c.Bytecode())
	vm.state.consts = bc.consts
	vm.state.nconsts = bc.nconsts
	vm.state.ndefs = bc.ndefs
	tvm := C.new_vm_with_state(C.CString(path), bc, vm.state)
	defer C.vm_dispose(tvm)
	if i := C.vm_run(tvm); i != 0 {
		C.go_vm_errorf(vm, C.CString("import error"))
	}
	vm.state = tvm.state

	mod := C.new_module()
	for name, sym := range c.Store {
		if sym.Scope == compiler.GlobalScope {
			o := vm.state.globals[sym.Index]
			// TODO: convert objects.

			if isExported(name) {
				C.module_set_exp(mod, C.CString(name), o)
			} else {
				C.module_set_unexp(mod, C.CString(name), o)
			}
		}
	}

	importTab[p] = mod
	vm.stack[vm.sp] = mod
	vm.sp++
}
