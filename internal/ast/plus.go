package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
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

func (p Plus) Eval() (obj.Object, error) {
	left, err := p.l.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	right, err := p.r.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType, obj.StringType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '+' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType, obj.StringType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '+' for type %v", right.Type())
	}

	switch {
	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		return obj.NewString(left.String() + right.String()), nil

	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		return obj.NewInteger(left.Int() + right.Int()), nil

	case obj.AssertTypes(left, obj.FloatType, obj.IntType) && obj.AssertTypes(right, obj.FloatType, obj.IntType):
		l, r := obj.ToFloat(left, right)
		return obj.NewFloat(l + r), nil

	default:
		return obj.NullObj, fmt.Errorf(
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
