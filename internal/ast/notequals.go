package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type NotEquals struct {
	l   Node
	r   Node
	pos int
}

func NewNotEquals(l, r Node, pos int) Node {
	return NotEquals{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (n NotEquals) Eval() (cobj.Object, error) {
	left, err := n.l.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	right, err := n.r.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	if !cobj.AssertTypes(left, cobj.IntType, cobj.FloatType, cobj.StringType, cobj.BoolType, cobj.NullType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '!=' for type %v", left.Type())
	}
	if !cobj.AssertTypes(right, cobj.IntType, cobj.FloatType, cobj.StringType, cobj.BoolType, cobj.NullType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '!=' for type %v", right.Type())
	}

	switch {
	case cobj.AssertTypes(right, cobj.BoolType, cobj.NullType) || cobj.AssertTypes(right, cobj.BoolType, cobj.NullType):
		return cobj.ParseBool(left.Int() != right.Int()), nil

	case cobj.AssertTypes(left, cobj.StringType) && cobj.AssertTypes(right, cobj.StringType):
		return cobj.ParseBool(left.String() != right.String()), nil

	case cobj.AssertTypes(left, cobj.IntType) && cobj.AssertTypes(right, cobj.IntType):
		return cobj.ParseBool(left.Int() != right.Int()), nil

	case cobj.AssertTypes(left, cobj.FloatType, cobj.IntType) && cobj.AssertTypes(right, cobj.FloatType, cobj.IntType):
		l, r := cobj.ToFloat(left, right)
		return cobj.ParseBool(l != r), nil
	default:
		return cobj.TrueObj, nil
	}
}

func (n NotEquals) String() string {
	return fmt.Sprintf("(%v != %v)", n.l, n.r)
}

func (n NotEquals) Compile(c *compiler.Compiler) (position int, err error) {
	if n.IsConstExpression() {
		o, err := n.Eval()
		if err != nil {
			return 0, c.NewError(n.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(n.pos)
		return position, err
	}

	if position, err = n.l.Compile(c); err != nil {
		return
	}
	if position, err = n.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpNotEqual)
	c.Bookmark(n.pos)
	return
}

func (n NotEquals) IsConstExpression() bool {
	return n.l.IsConstExpression() && n.r.IsConstExpression()
}
