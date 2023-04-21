package ast

import (
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type parseFn func(string, string) (Node, []error)

type Node interface {
	Eval() (cobj.Object, error)
	String() string
	compiler.Compilable
}

// Checks whether o is of type cobj.ErrorType.
func isError(o cobj.Object) bool {
	return o.Type() == cobj.ErrorType
}
