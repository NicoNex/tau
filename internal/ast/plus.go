package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type Plus struct {
	l   Node
	r   Node
	pos int
}

func NewPlus(l, r Node, pos int) Node {
	return Plus{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (p Plus) Eval() (cobj.Object, error) {
	left, err := p.l.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	right, err := p.r.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	if !cobj.AssertTypes(left, cobj.IntType, cobj.FloatType, cobj.StringType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '+' for type %v", left.Type())
	}
	if !cobj.AssertTypes(right, cobj.IntType, cobj.FloatType, cobj.StringType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '+' for type %v", right.Type())
	}

	switch {
	case cobj.AssertTypes(left, cobj.StringType) && cobj.AssertTypes(right, cobj.StringType):
		return cobj.NewString(left.String() + right.String()), nil

	case cobj.AssertTypes(left, cobj.IntType) && cobj.AssertTypes(right, cobj.IntType):
		return cobj.NewInteger(left.Int() + right.Int()), nil

	case cobj.AssertTypes(left, cobj.FloatType, cobj.IntType) && cobj.AssertTypes(right, cobj.FloatType, cobj.IntType):
		l, r := cobj.ToFloat(left, right)
		return cobj.NewFloat(l + r), nil

	default:
		return cobj.NullObj, fmt.Errorf(
			"invalid operation %v + %v (wrong types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}
}

func (p Plus) String() string {
	return fmt.Sprintf("(%v + %v)", p.l, p.r)
}

func (p Plus) Compile(c *compiler.Compiler) (position int, err error) {
	if p.IsConstExpression() {
		o, err := p.Eval()
		if err != nil {
			return 0, c.NewError(p.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(p.pos)
		return position, err
	}

	if position, err = p.l.Compile(c); err != nil {
		return
	}
	if position, err = p.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpAdd)
	c.Bookmark(p.pos)
	return
}

func (p Plus) IsConstExpression() bool {
	return p.l.IsConstExpression() && p.r.IsConstExpression()
}
