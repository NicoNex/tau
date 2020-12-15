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

func assertType(o obj.Object, types ...obj.Type) bool {
	for _, t := range types {
		if t == o.Type() {
			return true
		}
	}
	return false
}

func isFloat(o obj.Object) bool {
	return o.Type() == obj.FLOAT
}

func shouldConvert(l, r obj.Object) bool {
	return isFloat(l) || isFloat(r)
}

func convert(l, r obj.Object) (obj.Object, obj.Object) {
	if i, ok := l.(*obj.Integer); ok {
		l = obj.NewFloat(float64(*i))
	}
	if i, ok := r.(*obj.Integer); ok {
		r = obj.NewFloat(float64(*i))
	}
	return l, r
}

func toFloat(o obj.Object) obj.Object {
	if i, ok := o.(*obj.Integer); ok {
		return obj.NewFloat(float64(*i))
	} else if isFloat(o) {
		return o
	}
	return obj.NewError("cannot cast type %v to float", o.Type())
}
