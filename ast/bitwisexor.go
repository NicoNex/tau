package ast

import (
	"fmt"

	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type BitwiseXor struct {
	l Node
	r Node
}

func NewBitwiseXor(l, r Node) Node {
	return BitwiseXor{l, r}
}

func (b BitwiseXor) Eval(env *obj.Env) obj.Object {
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
		return obj.NewError("unsupported operator '^' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '^' for type %v", right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return obj.NewInteger(l ^ r)
}

func (b BitwiseXor) String() string {
	return fmt.Sprintf("(%v ^ %v)", b.l, b.r)
}

func (b BitwiseXor) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = b.l.Compile(c); err != nil {
		return
	}
	if position, err = b.r.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpBwXor), nil
}
