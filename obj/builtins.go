package obj

import (
	"fmt"
	"os"
	"io"
)

var Stdout io.Writer = os.Stdout

var Builtins = map[string]Builtin{
	"print": func(args ...Object) Object {
		var arguments []interface{}

		for _, a := range args {
			arguments = append(arguments, a.String())
		}
		fmt.Fprintln(Stdout, arguments...)
		return NullObj
	},
}
