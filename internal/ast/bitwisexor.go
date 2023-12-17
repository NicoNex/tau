package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseXor struct {
	l   Node
	r   Node
	pos int
}

func NewBitwiseXor(l, r Node, pos int) Node {
	return BitwiseXor{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (b BitwiseXor) Eval() (obj.Object, error) {
	left, err := b.l.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	right, err := b.r.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	if !obj.AssertTypes(left, obj.IntType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '^' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '^' for type %v", right.Type())
	}
	return obj.NewInteger(int64(left.(obj.Integer)) ^ int64(right.(obj.Integer))), nil
}

func (b BitwiseXor) String() string {
	return fmt.Sprintf("(%v ^ %v)", b.l, b.r)
}

func (b BitwiseXor) Compile(c *compiler.Compiler) (position int, err error) {
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
	position = c.Emit(code.OpBwXor)
	c.Bookmark(b.pos)
	return
}

func (b BitwiseXor) IsConstExpression() bool {
	return b.l.IsConstExpression() && b.r.IsConstExpression()
}
