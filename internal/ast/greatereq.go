package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type GreaterEq struct {
	l   Node
	r   Node
	pos int
}

func NewGreaterEq(l, r Node, pos int) Node {
	return GreaterEq{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (g GreaterEq) Eval() (obj.Object, error) {
	left, err := g.l.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	right, err := g.r.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	switch {
	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		return obj.ParseBool(int64(left.(obj.Integer)) >= int64(right.(obj.Integer))), nil

	case obj.AssertTypes(left, obj.IntType, obj.FloatType) && obj.AssertTypes(right, obj.IntType, obj.FloatType):
		l, r := obj.ToFloat(left, right)
		return obj.ParseBool(float64(l.(obj.Float)) >= float64(r.(obj.Float))), nil

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		return obj.ParseBool(left.String() >= right.String()), nil

	default:
		return obj.NullObj, fmt.Errorf("unsupported operator '>=' for types %v and %v", left.Type(), right.Type())
	}
}

func (g GreaterEq) String() string {
	return fmt.Sprintf("(%v >= %v)", g.l, g.r)
}

func (g GreaterEq) Compile(c *compiler.Compiler) (position int, err error) {
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
	position = c.Emit(code.OpGreaterThanEqual)
	c.Bookmark(g.pos)
	return
}

func (g GreaterEq) IsConstExpression() bool {
	return g.l.IsConstExpression() && g.r.IsConstExpression()
}
