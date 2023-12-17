package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Bang struct {
	n   Node
	pos int
}

func NewBang(n Node, pos int) Node {
	return Bang{
		n:   n,
		pos: pos,
	}
}

func (b Bang) Eval() (obj.Object, error) {
	value, err := b.n.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	switch value.Type() {
	case obj.BoolType:
		return obj.ParseBool(!obj.IsTruthy(value)), nil
	case obj.NullType:
		return obj.True, nil
	default:
		return obj.False, nil
	}
}

func (b Bang) String() string {
	return fmt.Sprintf("!%v", b.n)
}

func (b Bang) Compile(c *compiler.Compiler) (position int, err error) {
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
	position = c.Emit(code.OpBang)
	c.Bookmark(b.pos)
	return
}

func (b Bang) IsConstExpression() bool {
	return b.n.IsConstExpression()
}
