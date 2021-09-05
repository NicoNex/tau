package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type Times struct {
	l Node
	r Node
}

func NewTimes(l, r Node) Node {
	return Times{l, r}
}

func (t Times) Eval(env *obj.Env) obj.Object {
	var left = unwrap(t.l.Eval(env))
	var right = unwrap(t.r.Eval(env))

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '*' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '*' for type %v", right.Type())
	}

	if assertTypes(left, obj.INT) && assertTypes(right, obj.INT) {
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return obj.NewInteger(l * r)
	}

	left, right = toFloat(left, right)
	l := left.(*obj.Float).Val()
	r := right.(*obj.Float).Val()
	return obj.NewFloat(l * r)
}

func (t Times) String() string {
	return fmt.Sprintf("(%v * %v)", t.l, t.r)
}
