package ast

import (
	"fmt"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type Equals struct {
	l Node
	r Node
}

func NewEquals(l, r Node) Node {
	return Equals{l, r}
}

func (e Equals) Eval(env *obj.Env) obj.Object {
	var (
		left  = e.l.Eval(env)
		right = e.r.Eval(env)
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return obj.NewError("unsupported operator '==' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return obj.NewError("unsupported operator '==' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.BoolType, obj.NullType) || assertTypes(right, obj.BoolType, obj.NullType):
		return obj.ParseBool(left == right)

	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		return obj.ParseBool(l == r)

	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return obj.ParseBool(l == r)

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return obj.ParseBool(l == r)

	default:
		return obj.NewError(
			"invalid operation %v == %v (wrong types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}
}

func (e Equals) String() string {
	return fmt.Sprintf("(%v == %v)", e.l, e.r)
}

func (e Equals) Compile(c *compiler.Compiler) (position int) {
	return 0
}
