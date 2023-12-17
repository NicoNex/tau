package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Times struct {
	l   Node
	r   Node
	pos int
}

func NewTimes(l, r Node, pos int) Node {
	return Times{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (t Times) Eval() (obj.Object, error) {
	left, err := t.l.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	right, err := t.r.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '*' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '*' for type %v", right.Type())
	}

	if obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType) {
		return obj.NewInteger(int64(left.(obj.Integer)) * int64(right.(obj.Integer))), nil
	}

	l, r := obj.ToFloat(left, right)
	return obj.NewFloat(float64(l.(obj.Float)) * float64(r.(obj.Float))), nil
}

func (t Times) String() string {
	return fmt.Sprintf("(%v * %v)", t.l, t.r)
}

func (t Times) Compile(c *compiler.Compiler) (position int, err error) {
	if t.IsConstExpression() {
		o, err := t.Eval()
		if err != nil {
			return 0, c.NewError(t.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(t.pos)
		return position, err
	}

	if position, err = t.l.Compile(c); err != nil {
		return
	}
	if position, err = t.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpMul)
	c.Bookmark(t.pos)
	return
}

func (t Times) IsConstExpression() bool {
	return t.l.IsConstExpression() && t.r.IsConstExpression()
}
