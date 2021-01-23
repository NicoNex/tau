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
	if i, ok := l.(*obj.Integer); ok {
		l = obj.NewFloat(float64(*i))
	}
	if i, ok := r.(*obj.Integer); ok {
		r = obj.NewFloat(float64(*i))
	}
	return l, r
}
