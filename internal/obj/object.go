package obj

// #cgo CFLAGS: -Ofast -Ilibffi/include
// #cgo LDFLAGS: -Llibffi/lib ${SRCDIR}/libffi/lib/libffi.a -lm -L${SRCDIR}/../vm/bdwgc/lib ${SRCDIR}/../vm/bdwgc/lib/libgc.a
// #include <stdio.h>
// #include <stdlib.h>
// #include <stdint.h>
// #include "object.h"
//
// static inline uint32_t is_truthy(struct object o) {
// 	switch (o.type) {
// 	case obj_boolean:
// 		return o.data.i == 1;
// 	case obj_integer:
// 		return o.data.i != 0;
// 	case obj_float:
// 		return o.data.f != 0;
// 	case obj_null:
// 		return 0;
// 	default:
// 		return 1;
// 	}
// }
//
// static inline uint32_t is_error(struct object o) {
// 	return o.type == obj_error;
// }
//
// static inline char *error_msg(struct object err) {
// 	return err.data.str->str;
// }
//
// static inline int64_t int_val(struct object i) {
// 	return i.data.i;
// }
//
// static inline double float_val(struct object f) {
// 	return f.data.f;
// }
//
// static inline struct function *function_val(struct object fn) {
// 	return fn.data.fn;
// }
//
// static void set_stdout(int fd, const char *name) {
// #if !defined(_WIN32) && !defined(WIN32)
//	stdout = fdopen(fd, name);
// #endif
// }
import "C"

import (
	"errors"
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/tauerr"
)

type (
	Object           = C.struct_object
	Type             = C.enum_obj_type
	CompiledFunction = C.struct_function
)

const (
	NullType     Type = C.obj_null     // null
	BoolType          = C.obj_boolean  // bool
	IntType           = C.obj_integer  // int
	FloatType         = C.obj_float    // float
	BuiltinType       = C.obj_builtin  // builtin
	StringType        = C.obj_string   // string
	ErrorType         = C.obj_error    // error
	ListType          = C.obj_list     // list
	MapType           = C.obj_map      // map
	FunctionType      = C.obj_function // function
	ClosureType       = C.obj_closure  // closure
	ObjectType        = C.obj_object   // object
	PipeType          = C.obj_pipe     // pipe
	BytesType         = C.obj_bytes    // bytes
	NativeType        = C.obj_native   // native
)

var (
	Stdout io.Writer = os.Stdout
	Stdin  io.Reader = os.Stdin

	Builtins = [...]string{
		"len",
		"println",
		"print",
		"input",
		"string",
		"error",
		"type",
		"int",
		"float",
		"exit",
		"append",
		"new",
		"failed",
		"plugin",
		"pipe",
		"send",
		"recv",
		"close",
		"hex",
		"oct",
		"bin",
		"slice",
		"keys",
		"delete",
		"bytes",
	}

	NullObj  = C.null_obj
	TrueObj  = C.true_obj
	FalseObj = C.false_obj
)

func (o Object) Type() Type {
	return o._type
}

func (o Object) TypeString() string {
	return C.GoString(C.otype_str(o._type))
}

func (o Object) String() string {
	return C.GoString(C.object_str(o))
}

func (o Object) Int() int64 {
	return int64(C.int_val(o))
}

func (o Object) Float() float64 {
	return float64(C.float_val(o))
}

func (o Object) CompiledFunction() *CompiledFunction {
	return C.function_val(o)
}

func (cf CompiledFunction) Instructions() []byte {
	return C.GoBytes(unsafe.Pointer(cf.instructions), C.int(cf.len))
}

func (cf CompiledFunction) Len() int {
	return int(cf.len)
}

func (cf CompiledFunction) NumLocals() int {
	return int(cf.num_locals)
}

func (cf CompiledFunction) NumParams() int {
	return int(cf.num_params)
}

func (cf CompiledFunction) BKLen() int {
	return int(cf.bklen)
}

func ParseBool(b bool) Object {
	if b {
		return TrueObj
	}
	return FalseObj
}

func IsTruthy(o Object) bool {
	return C.is_truthy(o) == 1
}

func IsError(o Object) bool {
	return C.is_error(o) == 1
}

func GoError(o Object) error {
	if IsError(o) {
		return errors.New(C.GoString(C.error_msg(o)))
	}
	return nil
}

func NewBool(b bool) Object {
	if b {
		return C.true_obj
	} else {
		return C.false_obj
	}
}

func NewInteger(i int64) Object {
	return C.new_integer_obj(C.int64_t(i))
}

func NewFloat(f float64) Object {
	return C.new_float_obj(C.double(f))
}

func NewString(s string) Object {
	return C.new_string_obj(C.CString(s), C.size_t(len(s)))
}

func CArray[CT, GoT any](s []GoT) *CT {
	if len(s) > 0 {
		return (*CT)(unsafe.Pointer(&s[0]))
	}
	return nil
}

func NewFunctionCompiled(ins code.Instructions, nlocals, nparams int, bmarks []tauerr.Bookmark) Object {
	return C.new_function_obj(
		(*C.uchar)(unsafe.Pointer(&ins[0])),
		C.size_t(len(ins)),
		C.uint(nlocals),
		C.uint(nparams),
		CArray[C.struct_bookmark, tauerr.Bookmark](bmarks),
		C.uint(len(bmarks)),
	)
}

func AssertTypes(o Object, types ...Type) bool {
	for _, t := range types {
		if t == o.Type() {
			return true
		}
	}
	return false
}

func ToFloat(l, r Object) (left, right float64) {
	left, right = l.Float(), r.Float()

	if l.Type() == IntType {
		left = float64(l.Int())
	}
	if r.Type() == IntType {
		right = float64(r.Int())
	}
	return
}

func Println(a ...any) {
	fmt.Fprintln(Stdout, a...)
}

func Printf(s string, a ...any) {
	fmt.Fprintf(Stdout, s, a...)
}

func SetStdout(fd int, name string) {
	C.set_stdout(C.int(fd), C.CString(name))
}
