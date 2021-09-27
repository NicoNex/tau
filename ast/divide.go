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
	var (
		left  = unwrap(d.l.Eval(env))
		right = unwrap(d.r.Eval(env))
	)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '/' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '/' for type %v", right.Type())
	}

	left, right = toFloat(left, right)
	l := left.(*obj.Float).Val()
	r := right.(*obj.Float).Val()
	return obj.NewFloat(l / r)
}

func (d Divide) String() string {
	return fmt.Sprintf("(%v / %v)", d.l, d.r)
}
