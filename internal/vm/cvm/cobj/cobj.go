package cobj

// #include <stdlib.h>
// #include "../obj.h"
import "C"

import (
	"unsafe"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/tauerr"
)

type CObj = C.struct_object

func (c CObj) Type() obj.Type {
	return obj.Type(c._type)
}

func (c CObj) String() string {
	cstr := C.object_str(c)
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}

var (
	//extern null_obj
	NullObj CObj
	//extern true_obj
	TrueObj CObj
	//extern false_obj
	FalseObj CObj
)

func ParseBool(b bool) CObj {
	if b {
		return TrueObj
	}
	return FalseObj
}

func NewInteger(i int64) CObj {
	return C.new_integer_obj(C.int64_t(i))
}

func NewFloat(f float64) CObj {
	return C.new_float_obj(C.double(f))
}

func NewString(s string) CObj {
	return C.new_string_obj(C.CString(s), C.size_t(len(s)))
}

func NewFunctionCompiled(ins code.Instructions, nlocals, nparams int, bmarks []tauerr.Bookmark) CObj {
	return C.new_function_obj(
		(*C.uchar)(unsafe.Pointer(&ins[0])),
		C.size_t(len(ins)),
		C.uint(nlocals),
		C.uint(nparams),
		cBookmarks(bmarks),
		C.uint(len(bmarks)),
	)
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
