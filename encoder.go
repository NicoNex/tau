package tau

// #include "encoder.h"
// #include "internal/compiler/bytecode.h"
import "C"
import (
	"unsafe"

	"github.com/NicoNex/tau/internal/compiler"
)

func tauEncode(bcode compiler.Bytecode) []byte {
	buf := C.tau_encode(*(*C.struct_bytecode)(unsafe.Pointer(&bcode)))
	defer C.free_buffer(buf)
	return C.GoBytes(unsafe.Pointer(buf.buf), C.int(buf.len))
}

func tauDecode(b []byte) compiler.Bytecode {
	bc := C.tau_decode((*C.uint8_t)(unsafe.Pointer(&b[0])), C.size_t(len(b)))
	return *(*compiler.Bytecode)(unsafe.Pointer(&bc))
}
