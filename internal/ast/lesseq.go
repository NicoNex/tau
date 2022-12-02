package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type LessEq struct {
	l   Node
	r   Node
	pos int
}

func NewLessEq(l, r Node, pos int) Node {
	return LessEq{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (l LessEq) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(l.l.Eval(env))
		right = obj.Unwrap(l.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	switch {
	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return obj.ParseBool(l <= r)

	case obj.AssertTypes(left, obj.IntType, obj.FloatType) && obj.AssertTypes(right, obj.IntType, obj.FloatType):
		left, right = obj.ToFloat(left, right)
		l := left.(obj.Float)
		r := right.(obj.Float)
		return obj.ParseBool(l <= r)

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		l := left.(obj.String)
		r := right.(obj.String)
		return obj.ParseBool(l <= r)

	default:
		return obj.NewError("unsupported operator '<=' for types %v and %v", left.Type(), right.Type())
	}
}

func (l LessEq) String() string {
	return fmt.Sprintf("(%v <= %v)", l.l, l.r)
}

func (l LessEq) Compile(c *compiler.Compiler) (position int, err error) {
	if l.IsConstExpression() {
		o := l.Eval(nil)
		if e, ok := o.(obj.Error); ok {
			return 0, c.NewError(l.pos, string(e))
		}
		position = c.Emit(code.OpConstant, c.AddConstant(o))
		c.Bookmark(l.pos)
		return
	}

	if position, err = l.r.Compile(c); err != nil {
		return
	}
	if position, err = l.l.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpGreaterThanEqual)
	c.Bookmark(l.pos)
	return
}

func (l LessEq) IsConstExpression() bool {
	return l.l.IsConstExpression() && l.r.IsConstExpression()
}
