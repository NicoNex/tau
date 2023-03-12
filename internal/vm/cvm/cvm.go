package cvm

// #cgo CFLAGS: -Werror -Wall -g -Ofast -mtune=native -fopenmp
// #cgo LDFLAGS: -fopenmp
// #include <stdlib.h>
// #include "vm.h"
// #include "decoder.h"
// #include "obj.h"
import "C"
import (
	"unsafe"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/tauerr"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type CVM struct {
	vm *C.struct_vm
	bc *compiler.Bytecode
}

var vm CVM

// func New(file string, data []byte) CVM {
// 	d := (*C.uchar)(unsafe.Pointer(&data[0]))
// 	bcode := C.tau_decode(d, C.ulong(len(data)))
// 	vm = CVM{vm: C.new_vm(C.CString(file), bcode)}
// 	return vm
// }

func New(file string, bc *compiler.Bytecode) CVM {
	vm = CVM{vm: C.new_vm(C.CString(file), cbytecode(bc))}
	return vm
}

func (cvm CVM) Run() {
	C.vm_run(cvm.vm)
}

func cbytecode(bc *compiler.Bytecode) C.struct_bytecode {
	return C.struct_bytecode{
		insts:     (*C.uchar)(unsafe.Pointer(&bc.Instructions[0])),
		len:       C.size_t(len(bc.Instructions)),
		nconsts:   C.size_t(len(bc.Constants)),
		bookmarks: cBookmarks(bc.Bookmarks),
		bklen:     C.size_t(len(bc.Bookmarks)),
	}
}

func cObjs(objects []obj.Object) *C.struct_object {
	var objs = make([]C.struct_object, len(objects))

	for i, o := range objects {
		objs[i] = (C.struct_object)(cobj.ToC(o))
	}
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

// export parseAndCompile
// func parseAndCompile(cpath *C.char) C.struct_bytecode {
// 	path := C.GoString(cpath)

// 	p, err := obj.ImportLookup(path)
// 	if err != nil {
// 		msg := fmt.Sprintf("import: %v", err)
// 		C.vm_errorf(vm.vm, C.CString(msg))
// 	}

// 	b, err := os.ReadFile(p)
// 	if err != nil {
// 		msg := fmt.Sprintf("import: %v", err)
// 		C.vm_errorf(vm.vm, C.CString(msg))
// 	}

// 	tree, errs := parser.Parse(path, string(b))
// 	if len(errs) > 0 {
// 		m := fmt.Sprintf("import: multiple errors in module %s", path)
// 		// msg := string(parserError(p, errs))
// 		C.vm_errorf(vm.vm, C.CString(m))
// 	}

// 	c := compiler.New()
// 	c.SetFileInfo(path, string(b))
// 	if err := c.Compile(tree); err != nil {
// 		C.vm_errorf(vm.vm, C.CString(err.Error()))
// 	}
// }
