package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Less struct {
	l   Node
	r   Node
	pos int
}

func NewLess(l, r Node, pos int) Node {
	return Less{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (l Less) Eval() (obj.Object, error) {
	left, err := l.l.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	right, err := l.r.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	switch {
	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		return obj.ParseBool(int64(left.(obj.Integer)) < int64(right.(obj.Integer))), nil

	case obj.AssertTypes(left, obj.IntType, obj.FloatType) && obj.AssertTypes(right, obj.IntType, obj.FloatType):
		l, r := obj.ToFloat(left, right)
		return obj.ParseBool(float64(l.(obj.Float)) < float64(r.(obj.Float))), nil

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		return obj.ParseBool(left.String() < right.String()), nil

	default:
		return obj.NullObj, fmt.Errorf("unsupported operator '<' for types %v and %v", left.Type(), right.Type())
	}
}

func (l Less) String() string {
	return fmt.Sprintf("(%v < %v)", l.l, l.r)
}

func (l Less) Compile(c *compiler.Compiler) (position int, err error) {
	if l.IsConstExpression() {
		o, err := l.Eval()
		if err != nil {
			return 0, c.NewError(l.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(l.pos)
		return position, err
	}

	// the order of the compilation of the operands is inverted because we reuse
	// the code.OpGreaterThan OpCode.
	if position, err = l.r.Compile(c); err != nil {
		return
	}
	if position, err = l.l.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpGreaterThan)
	c.Bookmark(l.pos)
	return
}

func (l Less) IsConstExpression() bool {
	return l.l.IsConstExpression() && l.r.IsConstExpression()
}
