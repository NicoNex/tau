package ast

import (
	"fmt"
	"tau/obj"
)

type LessEq struct {
	l Node
	r Node
}

func NewLessEq(l, r Node) Node {
	return LessEq{l, r}
}

func (l LessEq) Eval() obj.Object {
	var left = l.l.Eval()
	var right = l.r.Eval()

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if left.Type() != right.Type() {
		return obj.NewError(
			"invalid operation %v <= %v (mismatched types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}

	switch left.Type() {
	case obj.INT:
		l := left.(*obj.Integer)
		r := right.(*obj.Integer)
		return obj.ParseBool(l.Val() <= r.Val())

	case obj.FLOAT:
		l := left.(*obj.Float)
		r := right.(*obj.Float)
		return obj.ParseBool(l.Val() <= r.Val())

	default:
		return obj.NewError("unsupported operator '<=' for type %v", left.Type())
	}
}

func (l LessEq) String() string {
	return fmt.Sprintf("(%v <= %v)", l.l, l.r)
}
