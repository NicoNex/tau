package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type Or struct {
	l   Node
	r   Node
	pos int
}

func NewOr(l, r Node, pos int) Node {
	return Or{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (o Or) Eval() (cobj.Object, error) {
	left, err := o.l.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	right, err := o.r.Eval()
	if err != nil {
		return cobj.NullObj, err
	}

	return cobj.ParseBool(cobj.IsTruthy(left) || cobj.IsTruthy(right)), nil
}

func (o Or) String() string {
	return fmt.Sprintf("(%v || %v)", o.l, o.r)
}

func (o Or) Compile(c *compiler.Compiler) (position int, err error) {
	if o.IsConstExpression() {
		object, err := o.Eval()
		if err != nil {
			return 0, c.NewError(o.pos, err.Error())
		}
		position = c.Emit(code.OpConstant, c.AddConstant(object))
		c.Bookmark(o.pos)
		return position, err
	}

	if position, err = o.l.Compile(c); err != nil {
		return
	}
	if position, err = o.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpOr)
	c.Bookmark(o.pos)
	return
}

func (o Or) IsConstExpression() bool {
	return o.l.IsConstExpression() && o.r.IsConstExpression()
}
