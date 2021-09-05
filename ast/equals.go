package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type Equals struct {
	l Node
	r Node
}

func NewEquals(l, r Node) Node {
	return Equals{l, r}
}

func (e Equals) Eval(env *obj.Env) obj.Object {
	var (
		left  = unwrap(e.l.Eval(env))
		right = unwrap(e.r.Eval(env))
	)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT, obj.FLOAT, obj.STRING) {
		return obj.NewError("unsupported operator '==' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT, obj.FLOAT, obj.STRING) {
		return obj.NewError("unsupported operator '==' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.STRING) && assertTypes(right, obj.STRING):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		return obj.ParseBool(l == r)

	case assertTypes(left, obj.INT) && assertTypes(right, obj.INT):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return obj.ParseBool(l == r)

	case assertTypes(left, obj.FLOAT, obj.INT) && assertTypes(right, obj.FLOAT, obj.INT):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return obj.ParseBool(l == r)

	default:
		return obj.NewError(
			"invalid operation %v == %v (wrong types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}
}

func (e Equals) String() string {
	return fmt.Sprintf("(%v == %v)", e.l, e.r)
}
