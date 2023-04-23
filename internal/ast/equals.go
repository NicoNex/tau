package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Equals struct {
	l   Node
	r   Node
	pos int
}

func NewEquals(l, r Node, pos int) Node {
	return Equals{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (e Equals) Eval() (obj.Object, error) {
	left, err := e.l.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	right, err := e.r.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '==' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '==' for type %v", right.Type())
	}

	switch {
	case obj.AssertTypes(left, obj.BoolType, obj.NullType) || obj.AssertTypes(right, obj.BoolType, obj.NullType):
		return obj.ParseBool(left.Int() == right.Int()), nil

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		return obj.ParseBool(left.String() == right.String()), nil

	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		return obj.ParseBool(left.Int() == right.Int()), nil

	case obj.AssertTypes(left, obj.FloatType, obj.IntType) && obj.AssertTypes(right, obj.FloatType, obj.IntType):
		l, r := obj.ToFloat(left, right)
		return obj.ParseBool(l == r), nil

	default:
		return obj.FalseObj, nil
	}
}

func (e Equals) String() string {
	return fmt.Sprintf("(%v == %v)", e.l, e.r)
}

func (e Equals) Compile(c *compiler.Compiler) (position int, err error) {
	if e.IsConstExpression() {
		o, err := e.Eval()
		if err != nil {
			return 0, c.NewError(e.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(e.pos)
		return position, err
	}

	if position, err = e.l.Compile(c); err != nil {
		return
	}
	if position, err = e.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpEqual)
	c.Bookmark(e.pos)
	return
}

func (e Equals) IsConstExpression() bool {
	return e.l.IsConstExpression() && e.r.IsConstExpression()
}
