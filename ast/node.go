package ast

import "tau/obj"

type Node interface {
	Eval() obj.Object
	String() string
}

var (
	NULL  = obj.NewNull()
	TRUE  = obj.NewBoolean(true)
	FALSE = obj.NewBoolean(false)
)

// Checks whether o is of type obj.ERROR.
func isError(o obj.Object) bool {
	return o.Type() == obj.ERROR
}

// Returns the internal object representation of the boolean b.
func btoo(b bool) obj.Object {
	if b {
		return TRUE
	}
	return FALSE
}
