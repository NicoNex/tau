package ast

import "github.com/NicoNex/tau/obj"

type Node interface {
	Eval(*obj.Env) obj.Object
	String() string
}

// Checks whether o is of type obj.ERROR.
func isError(o obj.Object) bool {
	return o.Type() == obj.ERROR
}
