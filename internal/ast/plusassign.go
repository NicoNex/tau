package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type PlusAssign struct {
	l   Node
	r   Node
	pos int
}

func NewPlusAssign(l, r Node, pos int) Node {
	return PlusAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (p PlusAssign) Eval(env *obj.Env) obj.Object {
	var (
		name  string
		left  = p.l.Eval(env)
		right = obj.Unwrap(p.r.Eval(env))
	)

	if ident, ok := p.l.(Identifier); ok {
		name = ident.String()
	}

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType, obj.StringType) {
		return obj.NewError("unsupported operator '+=' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType, obj.StringType) {
		return obj.NewError("unsupported operator '+=' for type %v", right.Type())
	}

	switch {
	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		if gs, ok := left.(obj.GetSetter); ok {
			l := gs.Object().(obj.String)
			r := right.(obj.String)
			return gs.Set(obj.String(l + r))
		}

		l := left.(obj.String)
		r := right.(obj.String)
		return env.Set(name, obj.String(l+r))

	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		if gs, ok := left.(obj.GetSetter); ok {
			l := gs.Object().(obj.Integer)
			r := right.(obj.Integer)
			return gs.Set(obj.Integer(l + r))
		}

		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return env.Set(name, obj.Integer(l+r))

	case obj.AssertTypes(left, obj.FloatType, obj.IntType) && obj.AssertTypes(right, obj.FloatType, obj.IntType):
		if gs, ok := left.(obj.GetSetter); ok {
			leftFl, rightFl := obj.ToFloat(gs.Object(), right)
			l := leftFl.(obj.Float)
			r := rightFl.(obj.Float)
			return gs.Set(obj.Float(l + r))
		}

		leftFl, rightFl := obj.ToFloat(left, right)
		l := leftFl.(obj.Float)
		r := rightFl.(obj.Float)
		return env.Set(name, obj.Float(l+r))

	default:
		return obj.NewError(
			"invalid operation %v += %v (wrong types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}
}

func (p PlusAssign) String() string {
	return fmt.Sprintf("(%v += %v)", p.l, p.r)
}

func (p PlusAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: p.l, r: Plus{l: p.l, r: p.r, pos: p.pos}, pos: p.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (p PlusAssign) IsConstExpression() bool {
	return false
}
