package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseOr struct {
	l   Node
	r   Node
	pos int
}

func NewBitwiseOr(l, r Node, pos int) Node {
	return BitwiseOr{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (b BitwiseOr) Eval(env *obj.Env) obj.Object {
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
		return obj.NewError("unsupported operator '|' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '|' for type %v", right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return obj.NewInteger(l | r)
}

func (b BitwiseOr) String() string {
	return fmt.Sprintf("(%v | %v)", b.l, b.r)
}

func (b BitwiseOr) Compile(c *compiler.Compiler) (position int, err error) {
	if b.IsConstExpression() {
		return c.Emit(code.OpConstant, c.AddConstant(b.Eval(nil))), nil
	}

	if position, err = b.l.Compile(c); err != nil {
		return
	}
	if position, err = b.r.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpBwOr), nil
}

func (b BitwiseOr) IsConstExpression() bool {
	return b.l.IsConstExpression() && b.r.IsConstExpression()
}
