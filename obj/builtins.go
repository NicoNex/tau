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
			return NewError("wrong number of arguments, expected 1, got %d", l)
		}
		return NewString(args[0].String())
	},
}
