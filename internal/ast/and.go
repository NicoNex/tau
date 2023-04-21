package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type And struct {
	l   Node
	r   Node
	pos int
}

func NewAnd(l, r Node, pos int) Node {
	return And{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (a And) Eval() (cobj.Object, error) {
	left, err := a.l.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	right, err := a.r.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	return cobj.ParseBool(cobj.IsTruthy(left) && cobj.IsTruthy(right)), nil
}

func (a And) String() string {
	return fmt.Sprintf("(%v && %v)", a.l, a.r)
}

func (a And) Compile(c *compiler.Compiler) (position int, err error) {
	if a.IsConstExpression() {
		o, err := a.Eval()
		if err != nil {
			return 0, c.NewError(a.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(a.pos)
		return position, err
	}

	if position, err = a.l.Compile(c); err != nil {
		return
	}
	if position, err = a.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpAnd)
	c.Bookmark(a.pos)
	return
}

func (a And) IsConstExpression() bool {
	return a.l.IsConstExpression() && a.r.IsConstExpression()
}
