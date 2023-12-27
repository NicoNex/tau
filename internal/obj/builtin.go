package obj

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

var (
	Stdout io.Writer      = os.Stdout
	Stdin  io.Reader      = os.Stdin
	Exit   func(code int) = os.Exit
)

func Println(a ...any) {
	fmt.Fprintln(Stdout, a...)
}

func Printf(s string, a ...any) {
	fmt.Fprintf(Stdout, s, a...)
}

type Builtin func(args ...Object) Object

func (b Builtin) Type() Type {
	return BuiltinType
}

func (b Builtin) String() string {
	return "<builtin function>"
}

func ResolveBuiltin(name string) (Builtin, bool) {
	for _, b := range Builtins {
		if name == b.Name {
			return b.Builtin, true
		}
	}
	return nil, false
}

type BuiltinImpl struct {
	Builtin Builtin
	Name    string
}

var Builtins = []BuiltinImpl{
	{
		Name: "len",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("len: wrong number of arguments, expected 1, got %d", l)
			}

			switch o := args[0].(type) {
			case List:
				return Integer(len(o))
			case String:
				return Integer(len(o))
			case Bytes:
				return Integer(len(o))
			default:
				return NewError("len: object of type %q has no length", o.Type())
			}
		},
	},
	{
		Name: "println",
		Builtin: func(args ...Object) Object {
			fmt.Fprintln(Stdout, toAnySlice(args)...)
			return NullObj
		},
	},
	{
		Name: "print",
		Builtin: func(args ...Object) Object {
			fmt.Fprint(Stdout, toAnySlice(args)...)
			return NullObj
		},
	},
	{
		Name: "input",
		Builtin: func(args ...Object) Object {
			var tmp string

			switch l := len(args); l {
			case 0:
				fmt.Scanln(&tmp)

			case 1:
				fmt.Print(args[0])
				fmt.Scanln(&tmp)

			default:
				return NewError("input: wrong number of arguments, expected 1, got %d", l)
			}
			return NewString(tmp)
		},
	},
	{
		Name: "string",
		Builtin: func(args ...Object) Object {
			if len(args) == 0 {
				return NewError("string: no argument provided")
			}

			if b, ok := args[0].(Bytes); ok {
				return String(b)
			}
			return NewString(fmt.Sprint(toAnySlice(args)...))
		},
	},
	{
		Name: "error",
		Builtin: func(args ...Object) Object {
			return NewError(fmt.Sprint(toAnySlice(args)...))
		},
	},
	{
		Name: "type",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("type: wrong number of arguments, expected 1, got %d", l)
			}
			return NewString(args[0].Type().String())
		},
	},
	{
		Name: "int",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("int: wrong number of arguments, expected 1, got %d", l)
			}

			switch o := args[0].(type) {
			case Integer:
				return Integer(o)

			case Float:
				return Integer(o)

			case String:
				if a, err := strconv.ParseInt(string(o), 10, 64); err == nil {
					return Integer(a)
				}
				return NewError("%v is not a number", args[0])

			default:
				return NewError("%v is not a number", args[0])
			}
		},
	},
	{
		Name: "float",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("float: wrong number of arguments, expected 1, got %d", l)
			}

			switch o := args[0].(type) {
			case Integer:
				return Float(o)

			case Float:
				return Float(o)

			case String:
				if a, err := strconv.ParseFloat(string(o), 64); err == nil {
					return Float(a)
				}
				return NewError("%v is not a number", args[0])

			default:
				return NewError("%v is not a number", args[0])
			}
		},
	},
	{
		Name: "exit",
		Builtin: func(args ...Object) Object {
			switch l := len(args); l {
			case 0:
				Exit(0)

			case 1:
				switch o := args[0].(type) {
				case Integer:
					Exit(int(o))

				case String, Error:
					fmt.Fprintln(Stdout, o)
					Exit(0)

				default:
					return NewError("exit: argument must be an integer, string or error")
				}

			case 2:
				msg, ok := args[0].(String)
				if !ok {
					return NewError("exit: first argument must be a string")
				}
				code, ok := args[1].(Integer)
				if !ok {
					return NewError("exit: second argument must be an int")
				}

				fmt.Fprintln(Stdout, string(msg))
				os.Exit(int(code))

			default:
				return NewError("exit: wrong number of arguments, max 2, got %d", l)
			}
			return NullObj
		},
	},
	{
		Name: "append",
		Builtin: func(args ...Object) Object {
			if len(args) == 0 {
				return NewError("append: no argument provided")
			}

			lst, ok := args[0].(List)
			if !ok {
				return NewError("append: first argument must be a list")
			}

			if len(args) > 1 {
				return append(lst, args[1:]...)
			}
			return lst
		},
	},
	{
		Name: "new",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 0 {
				return NewError("new: wrong number of arguments, expected 0, got %d", l)
			}
			return NewTauObject()
		},
	},
	{
		Name: "failed",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("failed: wrong number of arguments, expected 1, got %d", l)
			}

			_, ok := args[0].(Error)
			return ParseBool(ok)
		},
	},
	{
		Name: "plugin",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("plugin: wrong number of arguments, expected 1, got %d", l)
			}

			str, ok := args[0].(String)
			if !ok {
				return NewError("plugin: first argument must be a string, got %s instead", args[0].Type())
			}

			return NewNativePlugin(str.String())
		},
	},
	{
		Name: "pipe",
		Builtin: func(args ...Object) Object {
			switch l := len(args); l {
			case 0:
				return NewPipe()

			case 1:
				n, ok := args[0].(Integer)
				if !ok {
					return NewError("pipe: first argument must be an int, got %s instead", args[0].Type())
				}
				return NewPipeBuffered(int(n))

			default:
				return NewError("pipe: wrong number of arguments, expected 0 or 1, got %d", l)
			}
		},
	},
	{
		Name: "send",
		Builtin: func(args ...Object) (o Object) {
			if l := len(args); l != 2 {
				return NewError("send: wrong number of arguments, expected 2, got %d", l)
			}

			p, ok := args[0].(Pipe)
			if !ok {
				return NewError("send: first argument must be a pipe, got %s instead", args[0].Type())
			}

			p <- args[1]
			return args[1]
		},
	},
	{
		Name: "recv",
		Builtin: func(args ...Object) (o Object) {
			if l := len(args); l != 1 {
				return NewError("recv: wrong number of arguments, expected 1, got %d", l)
			}

			p, ok := args[0].(Pipe)
			if !ok {
				return NewError("recv: first argument must be a pipe, got %s instead", args[0].Type())
			}

			if ret := <-p; ret != nil {
				return ret
			}
			return NullObj
		},
	},
	{
		Name: "close",
		Builtin: func(args ...Object) (o Object) {
			if l := len(args); l != 1 {
				return NewError("close: wrong number of arguments, expected 1, got %d", l)
			}

			p, ok := args[0].(Pipe)
			if !ok {
				return NewError("close: first argument must be a pipe, got %s instead", args[0].Type())
			}

			close(p)
			return NullObj
		},
	},
	{
		Name: "hex",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("hex: wrong number of arguments, expected 1, got %d", l)
			}

			i, ok := args[0].(Integer)
			if !ok {
				return NewError("hex: first argument must be an int, got %s instead", args[0].Type())
			}

			return NewString(fmt.Sprintf("0x%x", i.Val()))
		},
	},
	{
		Name: "oct",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("oct: wrong number of arguments, expected 1, got %d", l)
			}

			i, ok := args[0].(Integer)
			if !ok {
				return NewError("oct: first argument must be an int, got %s instead", args[0].Type())
			}

			return NewString(fmt.Sprintf("%O", i.Val()))
		},
	},
	{
		Name: "bin",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("bin: wrong number of arguments, expected 1, got %d", l)
			}

			i, ok := args[0].(Integer)
			if !ok {
				return NewError("bin: first argument must be an int, got %s instead", args[0].Type())
			}

			return NewString(fmt.Sprintf("0b%b", i.Val()))
		},
	},
	{
		Name: "slice",
		Builtin: func(args ...Object) Object {
			if l := len(args); l != 3 {
				return NewError("slice: wrong number of arguments, expected 3, got %d", l)
			}

			s, ok := args[1].(Integer)
			if !ok {
				return NewError("slice: second argument must be an int, got %s instead", args[1].Type())
			}

			e, ok := args[2].(Integer)
			if !ok {
				return NewError("slice: third argument must be an int, got %s instead", args[2].Type())
			}

			var start, end = int(s), int(e)

			switch slice := args[0].(type) {
			case List:
				if start < 0 || end < 0 {
					return NewError("slice: invalid argument: index arguments must not be negative")
				} else if end > len(slice) {
					return NewError("slice: list bounds out of range %d with capacity %d", end, len(slice))
				}
				return slice[start:end]

			case String:
				if start < 0 || end < 0 {
					return NewError("slice: invalid argument: index arguments must not be negative")
				} else if end > len(slice) {
					return NewError("slice: string bounds out of range %d with capacity %d", end, len(slice))
				}
				return slice[start:end]

			case Bytes:
				if start < 0 || end < 0 {
					return NewError("slice: invalid argument: index arguments must not be negative")
				} else if end > len(slice) {
					return NewError("slice: bytes bounds out of range %d with capacity %d", end, len(slice))
				}
				return slice[start:end]

			default:
				return NewError("slice: first argument must be a list or string, got %s instead", args[0].Type())
			}
		},
	},
	{
		Name: "keys",
		Builtin: func(args ...Object) Object {
			if len(args) != 1 {
				return NewError("keys: wrong number of arguments, expected 1 got %d", len(args))
			}

			if m, ok := args[0].(Map); ok {
				ret := []Object{}

				for _, v := range m {
					ret = append(ret, v.Key)
				}
				return List(ret)
			}
			return NewError("keys: argument must be a map, got %v instead", args[0].Type())
		},
	},
	{
		Name: "delete",
		Builtin: func(args ...Object) Object {
			if len(args) != 2 {
				return NewError("delete: wrong number of arguments, expected 2 got %d", len(args))
			}

			if m, ok := args[0].(Map); ok {
				if h, ok := args[1].(Hashable); ok {
					delete(m, h.KeyHash())
					return NullObj
				}
				return NewError("delete: second argument must be one of boolean integer float string error, got %s instead", args[1].Type())
			}
			return NewError("delete: first argument must be a map, got %v instead", args[0].Type())
		},
	},
	{
		Name: "bytes",
		Builtin: func(args ...Object) Object {
			if len(args) != 1 {
				return NewError("bytes: expected 1 argument but got %d", len(args))
			}

			switch a := args[0].(type) {
			case String:
				return Bytes(a)
			case Integer:
				return make(Bytes, a)
			case List:
				ret := make(Bytes, len(a))
				for i, e := range a {
					integer, ok := e.(Integer)
					if !ok {
						return NewError("bytes: list cannot be converted to bytes")
					}
					ret[i] = byte(integer)
				}
				return ret
			default:
				return NewError("bytes: %s cannot be converted to bytes", a.Type())
			}
		},
	},
}

func toAnySlice(args []Object) []any {
	var ret = make([]any, len(args))
	for i, a := range args {
		ret[i] = a
	}
	return ret
}
