package obj

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var Stdout io.Writer = os.Stdout
var Stdin io.Reader = os.Stdin

var Builtins = map[string]Builtin{
	"println": func(args ...Object) Object {
		var arguments []interface{}

		for _, a := range args {
			if a.Type() == STRING {
				s := a.(*String)
				arguments = append(arguments, s.Val())
			} else {
				arguments = append(arguments, a.String())
			}
		}
		fmt.Fprintln(Stdout, arguments...)
		return NullObj
	},
	"print": func(args ...Object) Object {
		var arguments []interface{}

		for _, a := range args {
			if a.Type() == STRING {
				s := a.(*String)
				arguments = append(arguments, s.Val())
			} else {
				arguments = append(arguments, a.String())
			}
		}
		fmt.Fprint(Stdout, arguments...)
		return NullObj
	},
	"input": func(args ...Object) Object {
		var tmp string

		switch l := len(args); l {
		case 0:
			fmt.Scanln(&tmp)

		case 1:
			fmt.Print(args[0])
			fmt.Scanln(&tmp)

		default:
			return NewError("last: wrong number of arguments, expected 1, got %d", l)
		}
		return NewString(tmp)
	},
	"string": func(args ...Object) Object {
		var buf strings.Builder

		for _, a := range args {
			buf.WriteString(a.String())
		}
		return NewString(buf.String())
	},
	"int": func(args ...Object) Object {
		if l := len(args); l != 1 {
			return NewError("int: wrong number of arguments, expected 1, got %d", l)
		}

		switch o := args[0].(type) {
		case *Integer:
			return NewInteger(int64(*o))

		case *Float:
			return NewInteger(int64(*o))

		case *String:
			if a, err := strconv.ParseFloat(string(*o), 64); err == nil {
				return NewInteger(int64(a))
			}
			return NewError("%v is not a number", args[0])

		default:
			return NewError("%v is not a number", args[0])
		}
	},
	"float": func(args ...Object) Object {
		if l := len(args); l != 1 {
			return NewError("float: wrong number of arguments, expected 1, got %d", l)
		}

		switch o := args[0].(type) {
		case *Integer:
			return NewFloat(float64(*o))

		case *Float:
			return NewFloat(float64(*o))

		case *String:
			if a, err := strconv.ParseFloat(string(*o), 64); err == nil {
				return NewFloat(a)
			}
			return NewError("%v is not a number", args[0])

		default:
			return NewError("%v is not a number", args[0])
		}
	},
	"exit": func(args ...Object) Object {
		var l = len(args)

		if l == 0 {
			os.Exit(0)
		} else if l > 2 {
			return NewError("exit: wrong number of arguments, max 2, got %d", l)
		} else if l == 1 {
			switch o := args[0].(type) {
			case *Integer:
				os.Exit(int(*o))

			case *String:
				fmt.Fprintln(Stdout, o)
				os.Exit(0)

			default:
				return NewError("exit: argument must be an integer or string")
			}
		}

		msg, ok := args[0].(*String)
		if !ok {
			return NewError("exit: first argument must be a string")
		}

		code, ok := args[1].(*Integer)
		if !ok {
			return NewError("exit: second argument must be an int")
		}

		fmt.Fprintln(Stdout, string(*msg))
		os.Exit(int(*code))
		return NullObj
	},
	"append": func(args ...Object) Object {
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
	"push": func(args ...Object) Object {
		if len(args) == 0 {
			return NewError("push: no argument provided")
		}

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
	"len": func(args ...Object) Object {
		if l := len(args); l != 1 {
			return NewError("len: wrong number of arguments, expected 1, got %d", l)
		}

		switch o := args[0].(type) {
		case List:
			return NewInteger(int64(len(o)))
		case *String:
			return NewInteger(int64(len(*o)))
		default:
			return NewError("len: object of type %q has no length", o.Type())
		}
	},
	"range": func(args ...Object) Object {
		switch len(args) {
		case 1:
			if stop, ok := args[0].(*Integer); ok {
				return listify(0, int(*stop), 1)
			}
			return NewError("range: start value must be an int")

		case 2:
			start, ok := args[0].(*Integer)
			if !ok {
				return NewError("range: start value must be an int")
			}

			stop, ok := args[1].(*Integer)
			if !ok {
				return NewError("range: stop value must be an int")
			}
			return listify(int(*start), int(*stop), 1)

		case 3:
			start, ok := args[0].(*Integer)
			if !ok {
				return NewError("range: start value must be an int")
			}

			stop, ok := args[1].(*Integer)
			if !ok {
				return NewError("range: stop value must be an int")
			}

			step, ok := args[2].(*Integer)
			if !ok {
				return NewError("range: step value must be an int")
			}

			if s := int(*step); s != 0 {
				return listify(int(*start), int(*stop), s)
			}
			return NewError("range: step value must not be zero")

		default:
			return NewError("range: wrong number of arguments, max 3, got %d", len(args))
		}
	},
	"first": func(args ...Object) Object {
		if l := len(args); l != 1 {
			return NewError("first: wrong number of arguments, expected 1, got %d", l)
		}

		switch o := args[0].(type) {
		case List:
			return o[0]
		case *String:
			return NewString(string(string(*o)[0]))
		default:
			return NewError("first: wrong argument type, expected list, got %s", args[0].Type())
		}
	},
	"last": func(args ...Object) Object {
		if l := len(args); l != 1 {
			return NewError("last: wrong number of arguments, expected 1, got %d", l)
		}

		switch o := args[0].(type) {
		case List:
			return o[len(o)-1]
		case *String:
			s := string(*o)
			return NewString(string(s[len(s)-1]))
		default:
			return NewError("first: wrong argument type, expected list, got %s", args[0].Type())
		}
	},
	"tail": func(args ...Object) Object {
		if l := len(args); l != 1 {
			return NewError("tail: wrong number of arguments, expected 1, got %d", l)
		}

		switch o := args[0].(type) {
		case List:
			return o[1:]
		case *String:
			s := string(*o)
			return NewString(s[1:])
		default:
			return NewError("first: wrong argument type, expected list, got %s", args[0].Type())
		}
	},
}

func listify(start, stop, step int) List {
	var l List

	for i := start; i < stop; i += step {
		l = append(l, NewInteger(int64(i)))
	}
	return l
}
