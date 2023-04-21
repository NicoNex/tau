package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
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

func (e Equals) Eval() (cobj.Object, error) {
	left, err := e.l.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	right, err := e.r.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	if !cobj.AssertTypes(left, cobj.IntType, cobj.FloatType, cobj.StringType, cobj.BoolType, cobj.NullType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '==' for type %v", left.Type())
	}
	if !cobj.AssertTypes(right, cobj.IntType, cobj.FloatType, cobj.StringType, cobj.BoolType, cobj.NullType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '==' for type %v", right.Type())
	}

	switch {
	case cobj.AssertTypes(left, cobj.BoolType, cobj.NullType) || cobj.AssertTypes(right, cobj.BoolType, cobj.NullType):
		return cobj.ParseBool(left.Int() == right.Int()), nil

	case cobj.AssertTypes(left, cobj.StringType) && cobj.AssertTypes(right, cobj.StringType):
		return cobj.ParseBool(left.String() == right.String()), nil

	case cobj.AssertTypes(left, cobj.IntType) && cobj.AssertTypes(right, cobj.IntType):
		return cobj.ParseBool(left.Int() == right.Int()), nil

	case cobj.AssertTypes(left, cobj.FloatType, cobj.IntType) && cobj.AssertTypes(right, cobj.FloatType, cobj.IntType):
		l, r := cobj.ToFloat(left, right)
		return cobj.ParseBool(l == r), nil

	default:
		return cobj.FalseObj, nil
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
