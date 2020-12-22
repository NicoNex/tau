package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type GreaterEq struct {
	l Node
	r Node
}

func NewGreaterEq(l, r Node) Node {
	return GreaterEq{l, r}
}

func (g GreaterEq) Eval(env *obj.Env) obj.Object {
	var left = g.l.Eval(env)
	var right = g.r.Eval(env)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '>=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '>=' for type %v", right.Type())
	}

	if assertTypes(left, obj.INT) && assertTypes(right, obj.INT) {
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return obj.ParseBool(l >= r)
	}

	left, right = convert(left, right)
	l := left.(*obj.Float).Val()
	r := right.(*obj.Float).Val()
	return obj.ParseBool(l >= r)
}

func (g GreaterEq) String() string {
	return fmt.Sprintf("(%v >= %v)", g.l, g.r)
}
