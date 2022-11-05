package ast

import (
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Node interface {
	Eval(*obj.Env) obj.Object
	String() string
	compiler.Compilable
}

// Returns true if o needs to stop the execution flow.
func takesPrecedence(o obj.Object) bool {
	return isReturn(o) || isError(o) || isContinue(o) || isBreak(o)
}

func takesPrecedenceNoError(o obj.Object) bool {
	return isReturn(o) || isContinue(o) || isBreak(o)
}

// Checks whether o is of type obj.ErrorType.
func isError(o obj.Object) bool {
	return o.Type() == obj.ErrorType
}

// Checks wether o is of type obj.ReturnType.
func isReturn(o obj.Object) bool {
	return o.Type() == obj.ReturnType
}

// Checks whether o is a continue statement.
func isContinue(o obj.Object) bool {
	return o == obj.ContinueObj
}

// Checks whether o is a break statement.
func isBreak(o obj.Object) bool {
	return o == obj.BreakObj
}
