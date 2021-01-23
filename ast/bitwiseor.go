package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type BitwiseOr struct {
	l Node
	r Node
}

func NewBitwiseOr(l, r Node) Node {
	return BitwiseOr{l, r}
}

func (p BitwiseOr) Eval(env *obj.Env) obj.Object {
	var left = p.l.Eval(env)
	var right = p.r.Eval(env)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT) {
		return obj.NewError("unsupported operator '|' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT) {
		return obj.NewError("unsupported operator '|' for type %v", right.Type())
	}
	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return obj.NewInteger(l | r)
}

func (p BitwiseOr) String() string {
	return fmt.Sprintf("(%v | %v)", p.l, p.r)
}
