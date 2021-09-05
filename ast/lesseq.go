package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type LessEq struct {
	l Node
	r Node
}

func NewLessEq(l, r Node) Node {
	return LessEq{l, r}
}

func (l LessEq) Eval(env *obj.Env) obj.Object {
	var (
		left  = unwrap(l.l.Eval(env))
		right = unwrap(l.r.Eval(env))
	)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '<=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '<=' for type %v", right.Type())
	}

	if assertTypes(left, obj.INT) && assertTypes(right, obj.INT) {
		le := left.(*obj.Integer).Val()
		ri := right.(*obj.Integer).Val()
		return obj.ParseBool(le <= ri)
	}

	left, right = toFloat(left, right)
	le := left.(*obj.Float).Val()
	ri := right.(*obj.Float).Val()
	return obj.ParseBool(le <= ri)
}

func (l LessEq) String() string {
	return fmt.Sprintf("(%v <= %v)", l.l, l.r)
}
