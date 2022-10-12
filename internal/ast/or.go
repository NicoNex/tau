package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
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

func (o Or) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(o.l.Eval(env))
		right = obj.Unwrap(o.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	return obj.ParseBool(isTruthy(left) || isTruthy(right))
}

func (o Or) String() string {
	return fmt.Sprintf("(%v || %v)", o.l, o.r)
}

func (o Or) Compile(c *compiler.Compiler) (position int, err error) {
	if o.IsConstExpression() {
		position = c.Emit(code.OpConstant, c.AddConstant(o.Eval(nil)))
		c.Bookmark(o.pos)
		return
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
