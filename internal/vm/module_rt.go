//go:build taurt

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
	"unsafe"
)

// This is the module loader of a runtime built to run one bundled program. Its
// modules were compiled when the bundle was made and loaded before it started,
// so all an import has to do is find one. Nothing here reads a file or needs a
// parser or a compiler, which is what keeps them out of the binary.
//
//export vm_exec_load_module
func vm_exec_load_module(vm *C.struct_vm, cpath *C.char) int {
	path := C.GoString(cpath)

	if path == "" {
		C.go_vm_errorf(vm, C.CString("import: no file provided"))
		return 1
	}

	cp := C.CString(path)
	defer C.free(unsafe.Pointer(cp))

	var mod C.struct_object
	if C.modtab_get(vm.state.mods, cp, &mod) != 0 {
		vm.stack[vm.sp] = mod
		vm.sp++
		return 0
	}

	msg := fmt.Sprintf("import: no module named %q came with this program", path)
	C.go_vm_errorf(vm, C.CString(msg))
	return 1
}
