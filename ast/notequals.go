package ast

import (
	"fmt"
	"github.com/NicoNex/tau/obj"
)

type NotEquals struct {
	l Node
	r Node
}

func NewNotEquals(l, r Node) Node {
	return NotEquals{l, r}
}

func (n NotEquals) Eval(env *obj.Env) obj.Object {
	var left = n.l.Eval(env)
	var right = n.r.Eval(env)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT, obj.FLOAT, obj.STRING) {
		return obj.NewError("unsupported operator '!=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT, obj.FLOAT, obj.STRING) {
		return obj.NewError("unsupported operator '!=' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.STRING) && assertTypes(right, obj.STRING):
		l := left.(*obj.String)
		r := right.(*obj.String)
		return obj.ParseBool(l.Val() != r.Val())

	case assertTypes(left, obj.INT) && assertTypes(right, obj.INT):
		l := left.(*obj.Integer)
		r := right.(*obj.Integer)
		return obj.ParseBool(l.Val() != r.Val())

	case assertTypes(left, obj.FLOAT, obj.INT) && assertTypes(right, obj.FLOAT, obj.INT):
		left, right = convert(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return obj.ParseBool(l != r)

	default:
		return obj.NewError(
			"invalid operation %v != %v (wrong types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}
}

func (n NotEquals) String() string {
	return fmt.Sprintf("(%v != %v)", n.l, n.r)
}
