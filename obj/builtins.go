package obj

import "fmt"

var Builtins = map[string]Builtin{
	"println": func(args ...Object) Object {
		var arguments []interface{}

		for _, a := range args {
			arguments = append(arguments, a.String())
		}
		fmt.Println(arguments...)
		return NullObj
	},
}
