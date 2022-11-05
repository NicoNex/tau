package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type TimesAssign struct {
	l Node
	r Node
}

func NewTimesAssign(l, r Node) Node {
	return TimesAssign{l, r}
}

func (t TimesAssign) Eval(env *obj.Env) obj.Object {
	var (
		name  string
		left  = t.l.Eval(env)
		right = obj.Unwrap(t.r.Eval(env))
	)

	if ident, ok := t.l.(Identifier); ok {
		name = ident.String()
	}

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '*=' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '*=' for type %v", right.Type())
	}

	switch {
	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		if gs, ok := left.(obj.GetSetter); ok {
			l := gs.Object().(obj.Integer).Val()
			r := right.(obj.Integer).Val()
			return gs.Set(obj.NewInteger(l * r))
		}

		l := left.(obj.Integer).Val()
		r := right.(obj.Integer).Val()
		return env.Set(name, obj.NewInteger(l*r))

	case obj.AssertTypes(left, obj.FloatType, obj.IntType) && obj.AssertTypes(right, obj.FloatType, obj.IntType):
		if gs, ok := left.(obj.GetSetter); ok {
			leftFl, rightFl := obj.ToFloat(gs.Object(), right)
			l := leftFl.(obj.Float).Val()
			r := rightFl.(obj.Float).Val()
			return gs.Set(obj.NewFloat(l * r))
		}

		leftFl, rightFl := obj.ToFloat(left, right)
		l := leftFl.(obj.Float).Val()
		r := rightFl.(obj.Float).Val()
		return env.Set(name, obj.NewFloat(l*r))

	default:
		return obj.NewError(
			"invalid operation %v *= %v (wrong types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}
}

func (t TimesAssign) String() string {
	return fmt.Sprintf("(%v *= %v)", t.l, t.r)
}

func (t TimesAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{t.l, Times{t.l, t.r}}
	return n.Compile(c)
}

func (t TimesAssign) IsConstExpression() bool {
	return false
}
