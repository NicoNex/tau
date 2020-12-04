package ast

import (
	"fmt"
	"tau/obj"
)

type Divide struct {
	l Node
	r Node
}

func NewDivide(l, r Node) Node {
	return Divide{l, r}
}

func (d Divide) Eval() obj.Object {
	var left = d.l.Eval()
	var right = d.r.Eval()

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if left.Type() != right.Type() {
		return obj.NewError(
			"invalid operation %v / %v (mismatched types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}

	switch left.Type() {
	case obj.INT:
		l := left.(*obj.Integer)
		r := right.(*obj.Integer)
		return obj.NewInteger(l.Val() / r.Val())

	case obj.FLOAT:
		l := left.(*obj.Float)
		r := right.(*obj.Float)
		return obj.NewFloat(l.Val() / r.Val())

	default:
		return obj.NewError("unsupported operator '/' for type %v", left.Type())
	}
}

func (d Divide) String() string {
	return fmt.Sprintf("(%v / %v)", d.l, d.r)
}
