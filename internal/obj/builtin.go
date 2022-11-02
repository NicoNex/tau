package obj

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

var (
	Stdout io.Writer = os.Stdout
	Stdin  io.Reader = os.Stdin
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
	Name    string
	Builtin Builtin
}

var Builtins = []BuiltinImpl{
	{
		"len",
		func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("len: wrong number of arguments, expected 1, got %d", l)
			}

			switch o := Unwrap(args[0]).(type) {
			case List:
				return NewInteger(int64(len(o)))
			case String:
				return NewInteger(int64(len(o)))
			default:
				return NewError("len: object of type %q has no length", o.Type())
			}
		},
	},
	{
		"println",
		func(args ...Object) Object {
			fmt.Fprintln(Stdout, toAnySlice(args)...)
			return NullObj
		},
	},
	{
		"print",
		func(args ...Object) Object {
			fmt.Fprint(Stdout, toAnySlice(args)...)
			return NullObj
		},
	},
	{
		"input",
		func(args ...Object) Object {
			var tmp string

			args = UnwrapAll(args)
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
		"string",
		func(args ...Object) Object {
			if len(args) == 0 {
				return NewError("string: no argument provided")
			}

			args = UnwrapAll(args)
			if b, ok := args[0].(Bytes); ok {
				return String(b)
			}
			return NewString(fmt.Sprint(toAnySlice(args)...))
		},
	},
	{
		"error",
		func(args ...Object) Object {
			return NewError(fmt.Sprint(toAnySlice(args)...))
		},
	},
	{
		"type",
		func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("type: wrong number of arguments, expected 1, got %d", l)
			}
			return NewString(args[0].Type().String())
		},
	},
	{
		"int",
		func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("int: wrong number of arguments, expected 1, got %d", l)
			}

			args = UnwrapAll(args)
			switch o := args[0].(type) {
			case Integer:
				return NewInteger(int64(o))

			case Float:
				return NewInteger(int64(o))

			case String:
				if a, err := strconv.ParseInt(string(o), 10, 64); err == nil {
					return NewInteger(int64(a))
				}
				return NewError("%v is not a number", args[0])

			default:
				return NewError("%v is not a number", args[0])
			}
		},
	},
	{
		"float",
		func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("float: wrong number of arguments, expected 1, got %d", l)
			}

			args = UnwrapAll(args)
			switch o := args[0].(type) {
			case Integer:
				return NewFloat(float64(o))

			case Float:
				return NewFloat(float64(o))

			case String:
				if a, err := strconv.ParseFloat(string(o), 64); err == nil {
					return NewFloat(a)
				}
				return NewError("%v is not a number", args[0])

			default:
				return NewError("%v is not a number", args[0])
			}
		},
	},
	{
		"exit",
		func(args ...Object) Object {
			args = UnwrapAll(args)

			switch l := len(args); l {
			case 0:
				os.Exit(0)

			case 1:
				switch o := args[0].(type) {
				case Integer:
					os.Exit(int(o))

				case String, Error:
					fmt.Fprintln(Stdout, o)
					os.Exit(0)

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
		"append",
		func(args ...Object) Object {
			if len(args) == 0 {
				return NewError("append: no argument provided")
			}

			args = UnwrapAll(args)
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
		"push",
		func(args ...Object) Object {
			if len(args) == 0 {
				return NewError("push: no argument provided")
			}

			args = UnwrapAll(args)
			lst, ok := args[0].(List)
			if !ok {
				return NewError("push: first argument must be a list")
			}

			if len(args) > 1 {
				var tmp List

				for i := len(args) - 1; i > 0; i-- {
					tmp = append(tmp, args[i])
				}

				return append(tmp, lst...)
			}
			return lst
		},
	},
	{
		"range",
		func(args ...Object) Object {
			args = UnwrapAll(args)

			switch len(args) {
			case 1:
				if stop, ok := args[0].(Integer); ok {
					return listify(0, int(stop), 1)
				}
				return NewError("range: start value must be an int")

			case 2:
				start, ok := args[0].(Integer)
				if !ok {
					return NewError("range: start value must be an int")
				}

				stop, ok := args[1].(Integer)
				if !ok {
					return NewError("range: stop value must be an int")
				}
				return listify(int(start), int(stop), 1)

			case 3:
				start, ok := args[0].(Integer)
				if !ok {
					return NewError("range: start value must be an int")
				}

				stop, ok := args[1].(Integer)
				if !ok {
					return NewError("range: stop value must be an int")
				}

				step, ok := args[2].(Integer)
				if !ok {
					return NewError("range: step value must be an int")
				}

				if s := int(step); s != 0 {
					return listify(int(start), int(stop), s)
				}
				return NewError("range: step value must not be zero")

			default:
				return NewError("range: wrong number of arguments, max 3, got %d", len(args))
			}
		},
	},
	{
		"new",
		func(args ...Object) Object {
			if l := len(args); l != 0 {
				return NewError("new: wrong number of arguments, expected 0, got %d", l)
			}
			return NewClass()
		},
	},
	{
		"failed",
		func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("failed: wrong number of arguments, expected 1, got %d", l)
			}

			_, ok := Unwrap(args[0]).(Error)
			return ParseBool(ok)
		},
	},
	{
		"plugin",
		func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("plugin: wrong number of arguments, expected 1, got %d", l)
			}

			str, ok := Unwrap(args[0]).(String)
			if !ok {
				return NewError("plugin: first argument must be a string, got %s instead", args[0].Type())
			}

			return NewNativePlugin(str.String())
		},
	},
	{
		"pipe",
		func(args ...Object) Object {
			switch l := len(args); l {
			case 0:
				return NewPipe()

			case 1:
				n, ok := Unwrap(args[0]).(Integer)
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
		"send",
		func(args ...Object) (o Object) {
			if l := len(args); l != 2 {
				return NewError("send: wrong number of arguments, expected 2, got %d", l)
			}

			args = UnwrapAll(args)
			p, ok := args[0].(Pipe)
			if !ok {
				return NewError("send: first argument must be a pipe, got %s instead", args[0].Type())
			}

			p <- args[1]
			return args[1]
		},
	},
	{
		"recv",
		func(args ...Object) (o Object) {
			if l := len(args); l != 1 {
				return NewError("recv: wrong number of arguments, expected 1, got %d", l)
			}

			p, ok := Unwrap(args[0]).(Pipe)
			if !ok {
				return NewError("recv: first argument must be a pipe, got %s instead", args[0].Type())
			}

			if ret := Unwrap(<-p); ret != nil {
				return ret
			}
			return NullObj
		},
	},
	{
		"close",
		func(args ...Object) (o Object) {
			if l := len(args); l != 1 {
				return NewError("close: wrong number of arguments, expected 1, got %d", l)
			}

			p, ok := Unwrap(args[0]).(Pipe)
			if !ok {
				return NewError("close: first argument must be a pipe, got %s instead", args[0].Type())
			}

			close(p)
			return NullObj
		},
	},
	{
		"hex",
		func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("hex: wrong number of arguments, expected 1, got %d", l)
			}

			i, ok := Unwrap(args[0]).(Integer)
			if !ok {
				return NewError("hex: first argument must be an int, got %s instead", args[0].Type())
			}

			return NewString(fmt.Sprintf("0x%x", i.Val()))
		},
	},
	{
		"oct",
		func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("oct: wrong number of arguments, expected 1, got %d", l)
			}

			i, ok := Unwrap(args[0]).(Integer)
			if !ok {
				return NewError("oct: first argument must be an int, got %s instead", args[0].Type())
			}

			return NewString(fmt.Sprintf("%O", i.Val()))
		},
	},
	{
		"bin",
		func(args ...Object) Object {
			if l := len(args); l != 1 {
				return NewError("bin: wrong number of arguments, expected 1, got %d", l)
			}

			i, ok := Unwrap(args[0]).(Integer)
			if !ok {
				return NewError("bin: first argument must be an int, got %s instead", args[0].Type())
			}

			return NewString(fmt.Sprintf("0b%b", i.Val()))
		},
	},
	{
		"slice",
		func(args ...Object) Object {
			if l := len(args); l != 3 {
				return NewError("slice: wrong number of arguments, expected 3, got %d", l)
			}

			args = UnwrapAll(args)
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
		"open",
		func(args ...Object) Object {
			var l = len(args)

			if l != 1 && l != 2 {
				return NewError("open: wrong number of arguments, expected 1 or 2, got %d", l)
			}

			args = UnwrapAll(args)
			path, ok := args[0].(String)
			if !ok {
				return NewError("open: first argument must be a string, got %s instead", args[0].Type())
			}

			var flag = os.O_RDONLY
			if l == 2 {
				mode, ok := args[1].(String)
				if !ok {
					return NewError("open: second argument must be a string, got %s instead", args[1].Type())
				}
				parsed, err := parseFlag(string(mode))
				if err != nil {
					return NewError("open: %v", err)
				}
				flag = parsed
			}

			ret, err := NewFile(string(path), flag)
			if err != nil {
				return NewError("open: %v", err)
			}
			return ret
		},
	},
	{
		"bytes",
		func(args ...Object) Object {
			if len(args) != 1 {
				return NewError("bytes: expected 1 argument but got %d", len(args))
			}

			args = UnwrapAll(args)
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

func UnwrapAll(a []Object) []Object {
	for i, o := range a {
		a[i] = Unwrap(o)
	}
	return a
}

func listify(start, stop, step int) List {
	var l List

	for i := start; i < stop; i += step {
		l = append(l, NewInteger(int64(i)))
	}
	return l
}

func toAnySlice(args []Object) []any {
	var ret = make([]any, len(args))
	for i, a := range args {
		ret[i] = a
	}
	return ret
}

func parseFlag(f string) (int, error) {
	switch f {
	case "r":
		return os.O_RDONLY, nil
	case "w":
		return os.O_WRONLY | os.O_TRUNC | os.O_CREATE, nil
	case "a":
		return os.O_RDWR | os.O_APPEND | os.O_CREATE, nil
	case "x":
		return os.O_RDWR | os.O_CREATE | os.O_EXCL, nil
	case "rw":
		return os.O_RDWR | os.O_CREATE | os.O_TRUNC, nil
	default:
		return 0, fmt.Errorf("invalid file flag %q", f)
	}
}
