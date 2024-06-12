package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type And struct {
	l   Node
	r   Node
	pos int
}

func NewAnd(l, r Node, pos int) Node {
	return And{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (a And) Eval() (obj.Object, error) {
	left, err := a.l.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	right, err := a.r.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	return obj.ParseBool(obj.IsTruthy(left) && obj.IsTruthy(right)), nil
}

func (a And) String() string {
	return fmt.Sprintf("(%v && %v)", a.l, a.r)
}

func (a And) Compile(c *compiler.Compiler) (position int, err error) {
	if a.IsConstExpression() {
		o, err := a.Eval()
		if err != nil {
			return 0, c.NewError(a.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(a.pos)
		return position, err
	}

	if position, err = a.l.Compile(c); err != nil {
		return
	}

	jntPos := c.Emit(code.OpJumpNotTruthy, compiler.GenericPlaceholder)
	// Emit OpTrue because the value will be popped from the stack.
	position = c.Emit(code.OpTrue)
	position, err = a.r.Compile(c)
	if err != nil {
		return
	}
	position = c.Emit(code.OpAnd)
	jmpPos := c.Emit(code.OpJump, compiler.GenericPlaceholder)
	// Emit OpFalse because the expression needs to return false if jumped here.
	position = c.Emit(code.OpFalse)
	c.ReplaceOperand(jntPos, position)
	c.ReplaceOperand(jmpPos, c.Pos())
	c.Bookmark(a.pos)

	return
}

func (a And) IsConstExpression() bool {
	return a.l.IsConstExpression() && a.r.IsConstExpression()
}
