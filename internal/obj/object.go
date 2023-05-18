package obj

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
//	stdout = fdopen(fd, name);
// }
import "C"

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/tauerr"
)

type (
	Object           = C.struct_object
	Type             = C.enum_obj_type
	CompiledFunction = C.struct_function
)

//go:generate stringer -linecomment -type=Type
const (
	NullType      Type = C.obj_null      // null
	BoolType           = C.obj_boolean   // bool
	IntType            = C.obj_integer   // int
	FloatType          = C.obj_float     // float
	BuiltinType        = C.obj_builtin   // builtin
	StringType         = C.obj_string    // string
	ErrorType          = C.obj_error     // error
	ListType           = C.obj_list      // list
	MapType            = C.obj_map       // map
	FunctionType       = C.obj_function  // function
	ClosureType        = C.obj_closure   // closure
	ObjectType         = C.obj_object    // object
	PipeType           = C.obj_pipe      // pipe
	BytesType          = C.obj_bytes     // bytes
	GetsetterType      = C.obj_getsetter // getsetter
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
		"open",
		"bytes",
	}

	//extern null_obj
	NullObj Object
	//extern true_obj
	TrueObj Object
	//extern false_obj
	FalseObj Object

	ErrNoFileProvided = errors.New("no file provided")
)

func (o Object) Type() Type {
	return o._type
}

func (o Object) String() string {
	cstr := C.object_str(o)
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
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

func NewInteger(i int64) Object {
	return C.new_integer_obj(C.int64_t(i))
}

func NewFloat(f float64) Object {
	return C.new_float_obj(C.double(f))
}

func NewString(s string) Object {
	return C.new_string_obj(C.CString(s), C.size_t(len(s)))
}

func NewFunctionCompiled(ins code.Instructions, nlocals, nparams int, bmarks []tauerr.Bookmark) Object {
	return C.new_function_obj(
		(*C.uchar)(unsafe.Pointer(&ins[0])),
		C.size_t(len(ins)),
		C.uint(nlocals),
		C.uint(nparams),
		(*C.struct_bookmark)(unsafe.Pointer(&bmarks[0])),
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

func ImportLookup(taupath string) (string, error) {
	dir, file := filepath.Split(taupath)

	if file == "" {
		return "", ErrNoFileProvided
	}

	if filepath.Ext(file) == "" {
		file += ".tau"
	}

	path := filepath.Join(dir, file)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = filepath.Join("/lib", "tau", dir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return "", fmt.Errorf("%s: %w", path, err)
		}
	}

	return path, nil
}

func SetStdout(fd int, name string) {
	C.set_stdout(C.int(fd), C.CString(name))
}
