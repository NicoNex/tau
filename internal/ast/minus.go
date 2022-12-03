package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Minus struct {
	l   Node
	r   Node
	pos int
}

func NewMinus(l, r Node, pos int) Node {
	return Minus{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (m Minus) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(m.l.Eval(env))
		right = obj.Unwrap(m.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '-' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '-' for type %v", right.Type())
	}

	if obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType) {
		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return obj.Integer(l - r)
	}

	left, right = obj.ToFloat(left, right)
	l := left.(obj.Float)
	r := right.(obj.Float)
	return obj.Float(l - r)
}

func (m Minus) String() string {
	return fmt.Sprintf("(%v - %v)", m.l, m.r)
}

func (m Minus) Compile(c *compiler.Compiler) (position int, err error) {
	if m.IsConstExpression() {
		o := m.Eval(nil)
		if e, ok := o.(obj.Error); ok {
			return 0, c.NewError(m.pos, string(e))
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(m.pos)
		return
	}

	if position, err = m.l.Compile(c); err != nil {
		return
	}
	if position, err = m.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpSub)
	c.Bookmark(m.pos)
	return
}

func (m Minus) IsConstExpression() bool {
	return m.l.IsConstExpression() && m.r.IsConstExpression()
}
