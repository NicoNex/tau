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

func (t Times) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(t.l.Eval(env))
		right = obj.Unwrap(t.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '*' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '*' for type %v", right.Type())
	}

	if obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType) {
		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return obj.Integer(l * r)
	}

	left, right = obj.ToFloat(left, right)
	l := left.(obj.Float)
	r := right.(obj.Float)
	return obj.Float(l * r)
}

func (t Times) String() string {
	return fmt.Sprintf("(%v * %v)", t.l, t.r)
}

func (t Times) Compile(c *compiler.Compiler) (position int, err error) {
	if t.IsConstExpression() {
		o := t.Eval(nil)
		if e, ok := o.(obj.Error); ok {
			return 0, c.NewError(t.pos, string(e))
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(t.pos)
		return
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
