package ast

import (
	"fmt"
	"tau/obj"
)

type GreaterEq struct {
	l Node
	r Node
}

func NewGreaterEq(l, r Node) Node {
	return GreaterEq{l, r}
}

func (g GreaterEq) Eval() obj.Object {
	var left = g.l.Eval()
	var right = g.r.Eval()

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if left.Type() != right.Type() {
		return obj.NewError(
			"invalid operation %v >= %v (mismatched types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}

	switch left.Type() {
	case obj.INT:
		l := left.(*obj.Integer)
		r := right.(*obj.Integer)
		return obj.ParseBool(l.Val() >= r.Val())

	case obj.FLOAT:
		l := left.(*obj.Float)
		r := right.(*obj.Float)
		return obj.ParseBool(l.Val() >= r.Val())

	default:
		return obj.NewError("unsupported operator '>=' for type %v", left.Type())
	}
}

func (g GreaterEq) String() string {
	return fmt.Sprintf("(%v >= %v)", g.l, g.r)
}
