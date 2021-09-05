package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type And struct {
	l Node
	r Node
}

func NewAnd(l, r Node) Node {
	return And{l, r}
}

func (a And) Eval(env *obj.Env) obj.Object {
	var left = unwrap(a.l.Eval(env))
	var right = unwrap(a.r.Eval(env))

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	return obj.ParseBool(isTruthy(left) && isTruthy(right))
}

func (a And) String() string {
	return fmt.Sprintf("(%v && %v)", a.l, a.r)
}
