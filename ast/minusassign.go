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
		name        string
		isContainer bool
		left        = m.l.Eval(env)
		right       = unwrap(m.r.Eval(env))
	)

	if ident, ok := m.l.(Identifier); ok {
		name = ident.String()
	} else if _, isContainer = left.(*obj.Container); !isContainer {
		return obj.NewError("cannot assign to literal")
	}

	if isError(left) {
		return left
	}
	if isError(right) {
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
		l := unwrap(left).(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()

		if isContainer {
			return left.(*obj.Container).Set(obj.NewInteger(l - r))
		}
		return env.Set(name, obj.NewInteger(l-r))

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		leftFl, rightFl := toFloat(unwrap(left), right)
		l := leftFl.(*obj.Float).Val()
		r := rightFl.(*obj.Float).Val()

		if isContainer {
			return left.(*obj.Container).Set(obj.NewFloat(l - r))
		}
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
