package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
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

	if !assertTypes(left, obj.IntType, obj.FloatType, obj.StringType) {
		return obj.NewError("unsupported operator '+=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType, obj.StringType) {
		return obj.NewError("unsupported operator '+=' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		if gs, ok := left.(obj.GetSetter); ok {
			l := gs.Object().(*obj.String).Val()
			r := right.(*obj.String).Val()
			return gs.Set(obj.NewString(l + r))
		}

		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		return env.Set(name, obj.NewString(l+r))

	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		if gs, ok := left.(obj.GetSetter); ok {
			l := gs.Object().(*obj.Integer).Val()
			r := right.(*obj.Integer).Val()
			return gs.Set(obj.NewInteger(l + r))
		}

		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return env.Set(name, obj.NewInteger(l+r))

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		if gs, ok := left.(obj.GetSetter); ok {
			leftFl, rightFl := toFloat(gs.Object(), right)
			l := leftFl.(*obj.Float).Val()
			r := rightFl.(*obj.Float).Val()
			return gs.Set(obj.NewFloat(l + r))
		}

		leftFl, rightFl := toFloat(left, right)
		l := leftFl.(*obj.Float).Val()
		r := rightFl.(*obj.Float).Val()
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

func (p PlusAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{p.l, Plus{p.l, p.r}}
	return n.Compile(c)
}
