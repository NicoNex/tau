package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
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

func (p PrefixMinus) Eval() (cobj.Object, error) {
	right, err := p.n.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	switch right.Type() {
	case cobj.IntType:
		return cobj.NewInteger(-right.Int()), nil
	case cobj.FloatType:
		return cobj.NewFloat(-right.Float()), nil
	default:
		return cobj.NullObj, fmt.Errorf("unsupported prefix operator '-' for type %v", right.Type())
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
