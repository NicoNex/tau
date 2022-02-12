package ast

import (
	"fmt"

	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
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
	var (
		left  = obj.Unwrap(o.l.Eval(env))
		right = obj.Unwrap(o.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	return obj.ParseBool(isTruthy(left) || isTruthy(right))
}

func (o Or) String() string {
	return fmt.Sprintf("(%v || %v)", o.l, o.r)
}

func (o Or) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = o.l.Compile(c); err != nil {
		return
	}
	if position, err = o.r.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpOr), nil
}
