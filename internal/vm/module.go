//go:build !taurt

package vm

/*
#include <stdlib.h>
#include "vm.h"
#include "../obj/object.h"
#include "../compiler/bytecode.h"
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/parser"
)

// Importing a module that did not come with the program means reading it,
// parsing it and compiling it, which is why an interpreter carries a parser
// and a compiler. A runtime built for a bundled program carries neither: see
// module_rt.go.
//export vm_exec_load_module
func vm_exec_load_module(vm *C.struct_vm, cpath *C.char) int {
	path := C.GoString(cpath)

	if path == "" {
		C.go_vm_errorf(vm, C.CString("import: no file provided"))
		return 1
	}

	// A bundled module came with the program and was loaded before it started,
	// so it is looked for under the name it is imported with and the
	// filesystem is never touched: that is what makes a built program run on
	// its own.
	_, isBundled := bundled[path]

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
	cp := C.CString(p)
	defer C.free(unsafe.Pointer(cp))
	var mod C.struct_object
	if C.modtab_get(vm.state.mods, cp, &mod) != 0 {
		vm.stack[vm.sp] = mod
		vm.sp++
		return 0
	}

	if isBundled {
		// Loaded before the program ran, so finding it missing here means the
		// bundle is not the one it says it is.
		msg := fmt.Sprintf("import: %q came with the program but was not loaded", path)
		C.go_vm_errorf(vm, C.CString(msg))
		return 1
	}

	b, err := os.ReadFile(p)
	if err != nil {
		msg := fmt.Sprintf("import: %v", err)
		C.go_vm_errorf(vm, C.CString(msg))
		return 1
	}
	tree, errs := parser.Parse(p, string(b))
	if len(errs) > 0 {
		m := fmt.Sprintf("import: %v", errs[0])
		C.go_vm_errorf(vm, C.CString(m))
		return 1
	}

	c := compiler.NewImport(int(vm.state.ndefs), int(vm.state.consts.len))
	c.SetFileInfo(p, string(b))
	if err := c.Compile(tree); err != nil {
		C.go_vm_errorf(vm, C.CString(err.Error()))
		return 1
	}

	bc := c.Bytecode()
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

	mod = C.new_object()
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

	C.modtab_put(vm.state.mods, cp, mod)
	vm.stack[vm.sp] = mod
	vm.sp++
	return 0
}
