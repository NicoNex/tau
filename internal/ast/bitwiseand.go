package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseAnd struct {
	l Node
	r Node
}

func NewBitwiseAnd(l, r Node) Node {
	return BitwiseAnd{l, r}
}

func (b BitwiseAnd) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(b.l.Eval(env))
		right = obj.Unwrap(b.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType) {
		return obj.NewError("unsupported operator '&' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '&' for type %v", right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return obj.NewInteger(l & r)
}

func (b BitwiseAnd) String() string {
	return fmt.Sprintf("(%v & %v)", b.l, b.r)
}

func (b BitwiseAnd) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = b.l.Compile(c); err != nil {
		return
	}
	if position, err = b.r.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpBwAnd), nil
}
