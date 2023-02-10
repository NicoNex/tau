package cvm

// #cgo CFLAGS: -Werror -Wall -g -Ofast -mtune=native
// #include "vm.h"
// #include "decoder.h"
import "C"
import "unsafe"

type CVM struct {
	vm *C.struct_vm
}

func New(data []byte) CVM {
	d := (*C.uchar)(unsafe.Pointer(&data[0]))
	bcode := C.tau_decode(d, C.ulong(len(data)))
	return CVM{C.new_vm(bcode)}
}

func (cvm CVM) Run() {
	C.vm_run(cvm.vm)
	o := C.vm_last_popped_stack_elem(cvm.vm)
	C.print_obj(o)
}
