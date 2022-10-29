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

func isTruthy(o obj.Object) bool {
	switch val := o.(type) {
	case *obj.Boolean:
		return o == obj.True
	case *obj.Integer:
		return val.Val() != 0
	case *obj.Float:
		return val.Val() != 0
	case *obj.Null:
		return false
	default:
		return true
	}
}

func assertTypes(o obj.Object, types ...obj.Type) bool {
	for _, t := range types {
		if t == o.Type() {
			return true
		}
	}
	return false
}

func toFloat(l, r obj.Object) (obj.Object, obj.Object) {
	if i, ok := l.(obj.Integer); ok {
		l = obj.NewFloat(float64(*i))
	}
	if i, ok := r.(obj.Integer); ok {
		r = obj.NewFloat(float64(*i))
	}
	return l, r
}
