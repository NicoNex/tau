package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
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

func (d Divide) Eval() (cobj.Object, error) {

	left, err := d.l.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	right, err := d.r.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	if !cobj.AssertTypes(left, cobj.IntType, cobj.FloatType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '/' for type %v", left.Type())
	}
	if !cobj.AssertTypes(right, cobj.IntType, cobj.FloatType) {
		return cobj.NullObj, fmt.Errorf("unsupported operator '/' for type %v", right.Type())
	}

	l, r := cobj.ToFloat(left, right)
	return cobj.NewFloat(l / r), nil
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
