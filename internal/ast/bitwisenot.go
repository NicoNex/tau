package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseNot struct {
	n   Node
	pos int
}

func NewBitwiseNot(n Node, pos int) Node {
	return BitwiseNot{
		n:   n,
		pos: pos,
	}
}

func (b BitwiseNot) Eval() (obj.Object, error) {
	value, err := b.n.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	if !obj.AssertTypes(value, obj.IntType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '~' for type %v", value.Type())
	}

	return obj.NewInteger(^int64(value.(obj.Integer))), nil
}

func (b BitwiseNot) String() string {
	return fmt.Sprintf("~%v", b.n)
}

func (b BitwiseNot) Compile(c *compiler.Compiler) (position int, err error) {
	if b.IsConstExpression() {
		o, err := b.Eval()
		if err != nil {
			return 0, c.NewError(b.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(b.pos)
		return position, err
	}

	if position, err = b.n.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpBwNot)
	c.Bookmark(b.pos)
	return
}

func (b BitwiseNot) IsConstExpression() bool {
	return b.n.IsConstExpression()
}
