package vm

// #cgo CFLAGS: -Werror -Wall -Ifast -g -O3 -march=native -mtune=native
// #cgo LDFLAGS: -L${SRCDIR}/internal/vm/fast -L${SRCDIR}/internal/vm/fast/obj
// #include "vm.h"
// #include "decoder.h"
import "C"
import "unsafe"

type FastVM struct {
	vm *C.struct_vm
}

func NewFastVM(data []byte) FastVM {
	d := (*C.uchar)(unsafe.Pointer(&data[0]))
	bcode := C.tau_decode(d, C.ulong(len(data)))
	return FastVM{C.new_vm(bcode)}
}

func (f FastVM) Run() {
	C.vm_run(f.vm)
	o := C.vm_last_popped_stack_elem(f.vm)
	C.print_obj(o)
}
