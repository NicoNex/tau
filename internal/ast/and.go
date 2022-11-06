package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
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

func (a And) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(a.l.Eval(env))
		right = obj.Unwrap(a.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	return obj.ParseBool(obj.IsTruthy(left) && obj.IsTruthy(right))
}

func (a And) String() string {
	return fmt.Sprintf("(%v && %v)", a.l, a.r)
}

func (a And) Compile(c *compiler.Compiler) (position int, err error) {
	if a.IsConstExpression() {
		o := a.Eval(nil)
		if e, ok := o.(*obj.Error); ok {
			return 0, compiler.NewError(a.pos, string(*e))
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(a.pos)
		return
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
