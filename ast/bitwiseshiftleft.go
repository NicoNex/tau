package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type BitwiseLeftShift struct {
	l Node
	r Node
}

func NewBitwiseLeftShift(l, r Node) Node {
	return BitwiseLeftShift{l, r}
}

func (b BitwiseLeftShift) Eval(env *obj.Env) obj.Object {
	var (
		left  = unwrap(b.l.Eval(env))
		right = unwrap(b.r.Eval(env))
	)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT) {
		return obj.NewError("unsupported operator '<<' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT) {
		return obj.NewError("unsupported operator '<<' for type %v", right.Type())
	}
	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return obj.NewInteger(l << r)
}

func (b BitwiseLeftShift) String() string {
	return fmt.Sprintf("(%v << %v)", b.l, b.r)
}
