package ast

import (
	"fmt"
	"tau/obj"
)

type Plus struct {
	l Node
	r Node
}

func NewPlus(l, r Node) Node {
	return Plus{l, r}
}

func (p Plus) Eval() obj.Object {
	var left = p.l.Eval()
	var right = p.r.Eval()

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if left.Type() != right.Type() {
		return obj.NewError(
			"invalid operation %v + %v (mismatched types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}

	switch left.Type() {
	case obj.INT:
		l := left.(*obj.Integer)
		r := right.(*obj.Integer)
		return obj.NewInteger(l.Val() + r.Val())

	case obj.FLOAT:
		l := left.(*obj.Float)
		r := right.(*obj.Float)
		return obj.NewFloat(l.Val() + r.Val())

	case obj.STRING:
		l := left.(*obj.String)
		r := right.(*obj.String)
		return obj.NewString(l.Val() + r.Val())

	default:
		return obj.NewError("unsupported operator '+' for type %v", left.Type())
	}
}

func (p Plus) String() string {
	return fmt.Sprintf("(%v + %v)", p.l, p.r)
}
