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
	"strings"
	"unsafe"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/parser"
)

// cerrf hands the VM an error message and frees the copy C was given, which
// go_vm_errorf does not take over.
func cerrf(vm *C.struct_vm, format string, a ...any) {
	msg := C.CString(fmt.Sprintf(format, a...))
	C.go_vm_errorf(vm, msg)
	C.free(unsafe.Pointer(msg))
}

// The files whose import has not returned yet, innermost last. A file that
// shows up twice is a cycle, and without this the loader would follow it until
// the stack ran out.
//
// ponytail: a slice and a scan, imports nest a handful deep.
var inflight []string

// Importing a module that did not come with the program means reading it,
// parsing it and compiling it, which is why the interpreter carries a parser
// and a compiler. The runtime a bundled program is built on carries neither,
// and answers an import from what came with the program: internal/rt/rt.c.
//
//export vm_exec_load_module
func vm_exec_load_module(vm *C.struct_vm, cpath *C.char) int {
	path := C.GoString(cpath)

	if path == "" {
		cerrf(vm, "import: no file provided")
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
			cerrf(vm, "import: %v", err)
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
		cerrf(vm, "import: %q came with the program but was not loaded", path)
		return 1
	}

	// A file importing something that leads back to it would be read, compiled
	// and run again, one turn deeper every time, so the cycle is named instead.
	for _, f := range inflight {
		if f == p {
			cerrf(vm, "import: cycle %s -> %s", strings.Join(inflight, " -> "), p)
			return 1
		}
	}
	inflight = append(inflight, p)
	defer func() { inflight = inflight[:len(inflight)-1] }()

	b, err := os.ReadFile(p)
	if err != nil {
		cerrf(vm, "import: %v", err)
		return 1
	}
	tree, errs := parser.Parse(p, string(b))
	if len(errs) > 0 {
		cerrf(vm, "import: %v", errs[0])
		return 1
	}

	c := compiler.NewImport(int(vm.state.ndefs), int(vm.state.consts.len))
	c.SetFileInfo(p, string(b))
	if err := c.Compile(tree); err != nil {
		cerrf(vm, "%v", err)
		return 1
	}

	bc := c.Bytecode()
	// The resolved path, so that the modules imported by this one are looked
	// up next to it, and its own directory for the plugins it opens.
	setModuleDir(p)
	defer setModuleDir(C.GoString(vm.file))
	tvm := C.new_vm_with_state(C.CString(p), cbytecode(bc), vm.state)
	defer C.vm_dispose(tvm)
	// The VM that ran the module has already said what went wrong and where,
	// so there is nothing to add here.
	if C.vm_run(tvm) != 0 {
		return 1
	}
	vm.state = tvm.state

	mod = C.new_object()
	for name, sym := range c.Store {
		if sym.Scope == compiler.GlobalScope {
			o := C.get_global(vm.state.globals, C.size_t(sym.Index))

			// object_set copies the name, so the one C was given is ours to
			// free.
			if isExported(name) {
				cname := C.CString(name)
				if o._type == C.obj_object {
					C.object_set(mod, cname, C.object_to_module(o))
				} else {
					C.object_set(mod, cname, o)
				}
				C.free(unsafe.Pointer(cname))
			}
		}
	}

	C.modtab_put(vm.state.mods, cp, mod)
	vm.stack[vm.sp] = mod
	vm.sp++
	return 0
}
