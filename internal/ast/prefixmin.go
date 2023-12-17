package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type PrefixMinus struct {
	n   Node
	pos int
}

func NewPrefixMinus(n Node, pos int) Node {
	return PrefixMinus{
		n:   n,
		pos: pos,
	}
}

func (p PrefixMinus) Eval() (obj.Object, error) {
	right, err := p.n.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	switch r := right.(type) {
	case obj.Integer:
		return obj.NewInteger(-int64(r)), nil
	case obj.Float:
		return obj.NewFloat(-float64(r)), nil
	default:
		return obj.NullObj, fmt.Errorf("unsupported prefix operator '-' for type %v", right.Type())
	}
}

func (p PrefixMinus) String() string {
	return fmt.Sprintf("-%v", p.n)
}

func (p PrefixMinus) Compile(c *compiler.Compiler) (position int, err error) {
	if p.IsConstExpression() {
		o, err := p.Eval()
		if err != nil {
			return 0, c.NewError(p.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(p.pos)
		return position, err
	}

	if position, err = p.n.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpMinus)
	c.Bookmark(p.pos)
	return
}

func (p PrefixMinus) IsConstExpression() bool {
	return p.n.IsConstExpression()
}
