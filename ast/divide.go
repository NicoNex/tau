package ast

import (
	"fmt"
	"github.com/NicoNex/tau/obj"
)

type Divide struct {
	l Node
	r Node
}

func NewDivide(l, r Node) Node {
	return Divide{l, r}
}

func (d Divide) Eval(env *obj.Env) obj.Object {
	var left = d.l.Eval(env)
	var right = d.r.Eval(env)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '/' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '/' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.INT) && assertTypes(right, obj.INT):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return obj.NewInteger(l / r)

	default:
		left, right = convert(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return obj.NewFloat(l / r)
	}
}

func (d Divide) String() string {
	return fmt.Sprintf("(%v / %v)", d.l, d.r)
}
