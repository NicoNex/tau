package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type PlusAssign struct {
	l Node
	r Node
}

func NewPlusAssign(l, r Node) Node {
	return PlusAssign{l, r}
}

func (p PlusAssign) Eval(env *obj.Env) obj.Object {
	var (
		name        string
		isContainer bool
		left        = p.l.Eval(env)
		right       = unwrap(p.r.Eval(env))
	)

	if ident, ok := p.l.(Identifier); ok {
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

	if !assertTypes(left, obj.IntType, obj.FloatType, obj.StringType) {
		return obj.NewError("unsupported operator '+=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType, obj.StringType) {
		return obj.NewError("unsupported operator '+=' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := unwrap(left).(*obj.String).Val()
		r := right.(*obj.String).Val()
		if isContainer {
			return left.(*obj.Container).Set(obj.NewString(l + r))
		}
		return env.Set(name, obj.NewString(l+r))

	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := unwrap(left).(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		if isContainer {
			return left.(*obj.Container).Set(obj.NewInteger(l + r))
		}
		return env.Set(name, obj.NewInteger(l+r))

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		leftFl, rightFl := toFloat(unwrap(left), right)
		l := leftFl.(*obj.Float).Val()
		r := rightFl.(*obj.Float).Val()
		if isContainer {
			return left.(*obj.Container).Set(obj.NewFloat(l + r))
		}
		return env.Set(name, obj.NewFloat(l+r))

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
