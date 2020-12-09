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

func isTruthy(o obj.Object) bool {
	switch o.(type) {
	case *obj.Boolean:
		return o == obj.True
	case *obj.Null:
		return false
	default:
		return true
	}
}
