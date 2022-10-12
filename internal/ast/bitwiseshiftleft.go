package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseLeftShift struct {
	l   Node
	r   Node
	pos int
}

func NewBitwiseLeftShift(l, r Node, pos int) Node {
	return BitwiseLeftShift{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (b BitwiseLeftShift) Eval(env *obj.Env) obj.Object {
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

func (b BitwiseLeftShift) Compile(c *compiler.Compiler) (position int, err error) {
	if b.IsConstExpression() {
		position = c.Emit(code.OpConstant, c.AddConstant(b.Eval(nil)))
		c.Bookmark(b.pos)
		return
	}

	if position, err = b.l.Compile(c); err != nil {
		return
	}
	if position, err = b.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpBwLShift)
	c.Bookmark(b.pos)
	return
}

func (b BitwiseLeftShift) IsConstExpression() bool {
	return b.l.IsConstExpression() && b.r.IsConstExpression()
}
