package ast

import (
	"fmt"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
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
		right = t.r.Eval(env)
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

	if !assertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '*=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '*=' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return env.Set(name, obj.NewInteger(l*r))

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		leftFl, rightFl := toFloat(left, right)
		l := leftFl.(*obj.Float).Val()
		r := rightFl.(*obj.Float).Val()
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
	return 0, nil
}
