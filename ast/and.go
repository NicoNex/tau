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
	var (
		left  = unwrap(a.l.Eval(env))
		right = unwrap(a.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	return obj.ParseBool(isTruthy(left) && isTruthy(right))
}

func (a And) String() string {
	return fmt.Sprintf("(%v && %v)", a.l, a.r)
}
