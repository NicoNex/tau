package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type Greater struct {
	l   Node
	r   Node
	pos int
}

func NewGreater(l, r Node, pos int) Node {
	return Greater{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (g Greater) Eval() (cobj.Object, error) {
	left, err := g.l.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	right, err := g.r.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	switch {
	case cobj.AssertTypes(left, cobj.IntType) && cobj.AssertTypes(right, cobj.IntType):
		return cobj.ParseBool(left.Int() > right.Int()), nil

	case cobj.AssertTypes(left, cobj.IntType, cobj.FloatType) && cobj.AssertTypes(right, cobj.IntType, cobj.FloatType):
		l, r := cobj.ToFloat(left, right)
		return cobj.ParseBool(l > r), nil

	case cobj.AssertTypes(left, cobj.StringType) && cobj.AssertTypes(right, cobj.StringType):
		return cobj.ParseBool(left.String() > right.String()), nil

	default:
		return cobj.NullObj, fmt.Errorf("unsupported operator '>' for types %v and %v", left.Type(), right.Type())
	}
}

func (g Greater) String() string {
	return fmt.Sprintf("(%v > %v)", g.l, g.r)
}

func (g Greater) Compile(c *compiler.Compiler) (position int, err error) {
	if g.IsConstExpression() {
		o, err := g.Eval()
		if err != nil {
			return 0, c.NewError(g.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(g.pos)
		return position, err
	}

	if position, err = g.l.Compile(c); err != nil {
		return
	}
	if position, err = g.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpGreaterThan)
	c.Bookmark(g.pos)
	return
}

func (g Greater) IsConstExpression() bool {
	return g.l.IsConstExpression() && g.r.IsConstExpression()
}
