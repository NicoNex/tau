package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
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

func (t Times) Eval() (cobj.Object, error) {
	left, err := t.l.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	right, err := t.r.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	if !cobj.AssertTypes(left, cobj.IntType, cobj.FloatType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '*' for type %v", left.Type())
	}
	if !cobj.AssertTypes(right, cobj.IntType, cobj.FloatType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '*' for type %v", right.Type())
	}

	if cobj.AssertTypes(left, cobj.IntType) && cobj.AssertTypes(right, cobj.IntType) {
		return cobj.NewInteger(left.Int() * right.Int()), nil
	}

	l, r := cobj.ToFloat(left, right)
	return cobj.NewFloat(l * r), nil
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
