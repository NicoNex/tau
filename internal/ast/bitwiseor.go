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

func (b BitwiseOr) Eval() (obj.Object, error) {
	left, err := b.l.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	right, err := b.r.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	if !obj.AssertTypes(left, obj.IntType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '|' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '|' for type %v", right.Type())
	}
	return obj.NewInteger(left.Int() | right.Int()), nil
}

func (b BitwiseOr) String() string {
	return fmt.Sprintf("(%v | %v)", b.l, b.r)
}

func (b BitwiseOr) Compile(c *compiler.Compiler) (position int, err error) {
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
	position = c.Emit(code.OpBwOr)
	c.Bookmark(b.pos)
	return
}

func (b BitwiseOr) IsConstExpression() bool {
	return b.l.IsConstExpression() && b.r.IsConstExpression()
}
