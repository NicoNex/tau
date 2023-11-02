package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Divide struct {
	l   Node
	r   Node
	pos int
}

func NewDivide(l, r Node, pos int) Node {
	return Divide{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (d Divide) Eval() (obj.Object, error) {

	left, err := d.l.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	right, err := d.r.Eval()
	if err != nil {
		return obj.NullObj, err
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '/' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NullObj, fmt.Errorf("unsupported operator '/' for type %v", right.Type())
	}

	l, r := obj.ToFloat(left, right)
	return obj.NewFloat(l / r), nil
}

func (d Divide) String() string {
	return fmt.Sprintf("(%v / %v)", d.l, d.r)
}

func (d Divide) Compile(c *compiler.Compiler) (position int, err error) {
	if d.IsConstExpression() {
		o, err := d.Eval()
		if err != nil {
			return 0, c.NewError(d.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(d.pos)
		return position, err
	}

	if position, err = d.l.Compile(c); err != nil {
		return
	}
	if position, err = d.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpDiv)
	c.Bookmark(d.pos)
	return
}

func (d Divide) IsConstExpression() bool {
	return d.l.IsConstExpression() && d.r.IsConstExpression()
}
