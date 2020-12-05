package ast

import (
	"fmt"
	"tau/obj"
)

type NotEquals struct {
	l Node
	r Node
}

func NewNotEquals(l, r Node) Node {
	return NotEquals{l, r}
}

func (n NotEquals) Eval() obj.Object {
	var left = n.l.Eval()
	var right = n.r.Eval()

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if left.Type() != right.Type() {
		return obj.NewError(
			"invalid operation %v != %v (mismatched types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}

	switch left.Type() {
	case obj.INT:
		l := left.(*obj.Integer)
		r := right.(*obj.Integer)
		return obj.ParseBool(l.Val() != r.Val())

	case obj.FLOAT:
		l := left.(*obj.Float)
		r := right.(*obj.Float)
		return obj.ParseBool(l.Val() != r.Val())

	case obj.STRING:
		l := left.(*obj.String)
		r := right.(*obj.String)
		return obj.ParseBool(l.Val() != r.Val())

	case obj.BOOL:
		return obj.ParseBool(left != right)

	case obj.NULL:
		return obj.ParseBool(true)

	default:
		return obj.NewError("unsupported operator '!=' for type %v", left.Type())
	}
}

func (n NotEquals) String() string {
	return fmt.Sprintf("(%v != %v)", n.l, n.r)
}
