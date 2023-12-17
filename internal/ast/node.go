package ast

import (
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type parseFn func(string, string) (Node, []error)

type Node interface {
	Eval() (obj.Object, error)
	String() string
	compiler.Compilable
}

// Checks whether o is of type obj.ErrorType.
func isError(o obj.Object) bool {
	return o.Type() == obj.ErrorType
}
