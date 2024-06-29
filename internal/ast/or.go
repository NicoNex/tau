package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Or struct {
	l   Node
	r   Node
	pos int
}

func NewOr(l, r Node, pos int) Node {
	return Or{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (o Or) Eval() (obj.Object, error) {
	left, err := o.l.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	right, err := o.r.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	return obj.ParseBool(obj.IsTruthy(left) || obj.IsTruthy(right)), nil
}

func (o Or) String() string {
	return fmt.Sprintf("(%v || %v)", o.l, o.r)
}

func (o Or) Compile(c *compiler.Compiler) (position int, err error) {
	if o.IsConstExpression() {
		object, err := o.Eval()
		if err != nil {
			return 0, c.NewError(o.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(object))
		c.Bookmark(o.pos)
		return position, err
	}

	if position, err = o.l.Compile(c); err != nil {
		return
	}

	position = c.Emit(code.OpBang)
	jntPos := c.Emit(code.OpJumpNotTruthy, compiler.GenericPlaceholder)
	// Emit OpFalse because the value will be popped from the stack.
	position = c.Emit(code.OpFalse)

	if position, err = o.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpOr)
	jmpPos := c.Emit(code.OpJump, compiler.GenericPlaceholder)
	// Emit OpTrue because the expression needs to return false if jumped here.
	position = c.Emit(code.OpTrue)
	c.ReplaceOperand(jntPos, position)
	c.ReplaceOperand(jmpPos, c.Pos())
	c.Bookmark(o.pos)

	return
}

func (o Or) IsConstExpression() bool {
	return o.l.IsConstExpression() && o.r.IsConstExpression()
}
