package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type BitwiseRightShift struct {
	l   Node
	r   Node
	pos int
}

func NewBitwiseRightShift(l, r Node, pos int) Node {
	return BitwiseRightShift{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (b BitwiseRightShift) Eval() (cobj.Object, error) {
	left, err := b.l.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	right, err := b.r.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	if !cobj.AssertTypes(left, cobj.IntType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '>>' for type %v", left.Type())
	}
	if !cobj.AssertTypes(right, cobj.IntType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '>>' for type %v", right.Type())
	}
	return cobj.NewInteger(left.Int() >> right.Int()), nil
}

func (b BitwiseRightShift) String() string {
	return fmt.Sprintf("(%v >> %v)", b.l, b.r)
}

func (b BitwiseRightShift) Compile(c *compiler.Compiler) (position int, err error) {
	if b.IsConstExpression() {
		o, err := b.Eval()
		if err != nil {
			return 0, c.NewError(b.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(b.pos)
		return position, err
	}

	if position, err = b.l.Compile(c); err != nil {
		return
	}
	if position, err = b.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpBwRShift)
	c.Bookmark(b.pos)
	return
}

func (b BitwiseRightShift) IsConstExpression() bool {
	return b.l.IsConstExpression() && b.r.IsConstExpression()
}
