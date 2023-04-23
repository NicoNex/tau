package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
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

func (n NotEquals) Eval() (obj.Object, error) {
	left, err := n.l.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	right, err := n.r.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '!=' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '!=' for type %v", right.Type())
	}

	switch {
	case obj.AssertTypes(right, obj.BoolType, obj.NullType) || obj.AssertTypes(right, obj.BoolType, obj.NullType):
		return obj.ParseBool(left.Int() != right.Int()), nil

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		return obj.ParseBool(left.String() != right.String()), nil

	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		return obj.ParseBool(left.Int() != right.Int()), nil

	case obj.AssertTypes(left, obj.FloatType, obj.IntType) && obj.AssertTypes(right, obj.FloatType, obj.IntType):
		l, r := obj.ToFloat(left, right)
		return obj.ParseBool(l != r), nil
	default:
		return obj.TrueObj, nil
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
