package ast

import (
	"fmt"
	"tau/obj"
)

type Minus struct {
	l Node
	r Node
}

func NewMinus(l, r Node) Node {
	return Minus{l, r}
}

func (m Minus) Eval(env *obj.Env) obj.Object {
	var left = m.l.Eval(env)
	var right = m.r.Eval(env)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if left.Type() != right.Type() {
		return obj.NewError(
			"invalid operation %v - %v (mismatched types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}

	switch left.Type() {
	case obj.INT:
		l := left.(*obj.Integer)
		r := right.(*obj.Integer)
		return obj.NewInteger(l.Val() - r.Val())

	case obj.FLOAT:
		l := left.(*obj.Float)
		r := right.(*obj.Float)
		return obj.NewFloat(l.Val() - r.Val())

	default:
		return obj.NewError("unsupported operator '-' for type %v", left.Type())
	}
}

func (m Minus) String() string {
	return fmt.Sprintf("(%v - %v)", m.l, m.r)
}
