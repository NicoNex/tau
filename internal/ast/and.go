package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
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
		left  = obj.Unwrap(a.l.Eval(env))
		right = obj.Unwrap(a.r.Eval(env))
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

func (a And) Compile(c *compiler.Compiler) (position int, err error) {
	if a.IsConstExpression() {
		return c.Emit(code.OpConstant, c.AddConstant(a.Eval(nil))), nil
	}

	if position, err = a.l.Compile(c); err != nil {
		return
	}
	if position, err = a.r.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpAnd), nil
}

func (a And) IsConstExpression() bool {
	return a.l.IsConstExpression() && a.r.IsConstExpression()
}
