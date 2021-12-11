package ast

import (
	"fmt"

	"github.com/NicoNex/tau/compiler"
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
		left  = b.l.Eval(env)
		right = b.r.Eval(env)
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType) {
		return obj.NewError("unsupported operator '<<' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '<<' for type %v", right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return obj.NewInteger(l << r)
}

func (b BitwiseLeftShift) String() string {
	return fmt.Sprintf("(%v << %v)", b.l, b.r)
}

func (b BitwiseLeftShift) Compile(c *compiler.Compiler) int {
	return 0
}
