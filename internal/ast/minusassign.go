package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type MinusAssign struct {
	l Node
	r Node
}

func NewMinusAssign(l, r Node) Node {
	return MinusAssign{l, r}
}

func (m MinusAssign) Eval(env *obj.Env) obj.Object {
	var (
		name  string
		left  = m.l.Eval(env)
		right = obj.Unwrap(m.r.Eval(env))
	)

	if ident, ok := m.l.(Identifier); ok {
		name = ident.String()
	}

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '-=' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '-=' for type %v", right.Type())
	}

	switch {
	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		if gs, ok := left.(obj.GetSetter); ok {
			l := gs.Object().(obj.Integer)
			r := right.(obj.Integer)
			return gs.Set(obj.Integer(l - r))
		}

		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return env.Set(name, obj.Integer(l-r))

	case obj.AssertTypes(left, obj.FloatType, obj.IntType) && obj.AssertTypes(right, obj.FloatType, obj.IntType):
		if gs, ok := left.(obj.GetSetter); ok {
			leftFl, rightFl := obj.ToFloat(gs.Object(), right)
			l := leftFl.(obj.Float)
			r := rightFl.(obj.Float)
			return gs.Set(obj.Float(l - r))
		}

		leftFl, rightFl := obj.ToFloat(left, right)
		l := leftFl.(obj.Float)
		r := rightFl.(obj.Float)
		return env.Set(name, obj.Float(l-r))

	default:
		return obj.NewError(
			"invalid operation %v -= %v (wrong types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}
}

func (m MinusAssign) String() string {
	return fmt.Sprintf("(%v -= %v)", m.l, m.r)
}

func (m MinusAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{m.l, Minus{m.l, m.r}}
	return n.Compile(c)
}

func (m MinusAssign) IsConstExpression() bool {
	return false
}
