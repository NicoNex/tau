package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type Or struct {
	l Node
	r Node
}

func NewOr(l, r Node) Node {
	return Or{l, r}
}

func (o Or) Eval(env *obj.Env) obj.Object {
	var left = unwrap(o.l.Eval(env))
	var right = unwrap(o.r.Eval(env))

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	return obj.ParseBool(isTruthy(left) || isTruthy(right))
}

func (o Or) String() string {
	return fmt.Sprintf("(%v || %v)", o.l, o.r)
}
