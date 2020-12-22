package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type Plus struct {
	l Node
	r Node
}

func NewPlus(l, r Node) Node {
	return Plus{l, r}
}

func (p Plus) Eval(env *obj.Env) obj.Object {
	var left = p.l.Eval(env)
	var right = p.r.Eval(env)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT, obj.FLOAT, obj.STRING) {
		return obj.NewError("unsupported operator '+' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT, obj.FLOAT, obj.STRING) {
		return obj.NewError("unsupported operator '+' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.STRING) && assertTypes(right, obj.STRING):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		return obj.NewString(l + r)

	case assertTypes(left, obj.INT) && assertTypes(right, obj.INT):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return obj.NewInteger(l + r)

	case assertTypes(left, obj.FLOAT, obj.INT) && assertTypes(right, obj.FLOAT, obj.INT):
		left, right = convert(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return obj.NewFloat(l + r)

	default:
		return obj.NewError(
			"invalid operation %v + %v (wrong types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}
}

func (p Plus) String() string {
	return fmt.Sprintf("(%v + %v)", p.l, p.r)
}
