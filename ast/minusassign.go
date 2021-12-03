package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
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
		right = m.r.Eval(env)
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

	if !assertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '-=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '-=' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return env.Set(name, obj.NewInteger(l-r))

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		leftFl, rightFl := toFloat(left, right)
		l := leftFl.(*obj.Float).Val()
		r := rightFl.(*obj.Float).Val()
		return env.Set(name, obj.NewFloat(l-r))

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
