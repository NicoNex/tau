package ast

import (
	"fmt"
	"tau/obj"
)

type Equals struct {
	l Node
	r Node
}

func NewEquals(l, r Node) Node {
	return Equals{l, r}
}

func (e Equals) Eval() obj.Object {
	var left = e.l.Eval()
	var right = e.r.Eval()

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if left.Type() != right.Type() {
		return obj.NewError(
			"invalid operation %v == %v (mismatched types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}

	switch left.Type() {
	case obj.INT:
		l := left.(*obj.Integer)
		r := right.(*obj.Integer)
		return btoo(l.Val() == r.Val())

	case obj.FLOAT:
		l := left.(*obj.Float)
		r := right.(*obj.Float)
		return btoo(l.Val() == r.Val())

	case obj.BOOL:
		return btoo(left == right)

	case obj.NULL:
		return btoo(true)

	default:
		return obj.NewError("unsupported operator '==' for type %v", left.Type())
	}
}

func (e Equals) String() string {
	return fmt.Sprintf("(%v == %v)", e.l, e.r)
}
