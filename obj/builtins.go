package obj

import (
	"fmt"
	"io"
	"os"
)

var Stdout io.Writer = os.Stdout

var Builtins = map[string]Builtin{
	"println": func(args ...Object) Object {
		var arguments []interface{}

		for _, a := range args {
			arguments = append(arguments, a.String())
		}
		fmt.Fprintln(Stdout, arguments...)
		return NullObj
	},
	"print": func(args ...Object) Object {
		var arguments []interface{}

		for _, a := range args {
			arguments = append(arguments, a.String())
		}
		fmt.Fprint(Stdout, arguments...)
		return NullObj
	},
	"string": func(args ...Object) Object {
		if l := len(args); l != 1 {
			return NewError("string: wrong number of arguments, expected 1, got %d", l)
		}

		if s, ok := args[0].(*String); ok {
			return NewString(string(*s))
		}
		return NewString(args[0].String())
	},
	"int": func(args ...Object) Object {
		if l := len(args); l != 1 {
			return NewError("string: wrong number of arguments, expected 1, got %d", l)
		}

		if i, ok := args[0].(*Integer); ok {
			return NewInteger(int64(*i))
		}
		if a, b := ObjectToInt(args[0]); b {
			return NewInteger(a)
		}
		return NewError("Not an integer")
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

		if list, ok := args[0].(List); ok {
			if len(list) > 0 {
				return list[0]
			}
			return NullObj
		}
		return NewError("first: wrong argument type, expected list, got %s", args[0].Type())
	},
	"last": func(args ...Object) Object {
		if l := len(args); l != 1 {
			return NewError("last: wrong number of arguments, expected 1, got %d", l)
		}

		if list, ok := args[0].(List); ok {
			if len(list) > 0 {
				return list[len(list)-1]
			}
			return NullObj
		}
		return NewError("last: wrong argument type, expected list, got %s", args[0].Type())
	},
	"tail": func(args ...Object) Object {
		if l := len(args); l != 1 {
			return NewError("tail: wrong number of arguments, expected 1, got %d", l)
		}

		if list, ok := args[0].(List); ok {
			if len(list) > 0 {
				return list[1:]
			}
			return NullObj
		}
		return NewError("tail: wrong argument type, expected list, got %s", args[0].Type())
	},
}

func listify(start, stop, step int) List {
	var l List

	for i := start; i < stop; i += step {
		l = append(l, NewInteger(int64(i)))
	}
	return l
}
