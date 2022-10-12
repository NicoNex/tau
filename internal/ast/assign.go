package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Assign struct {
	l Node
	r Node
}

func NewAssign(l, r Node) Node {
	return Assign{l, r}
}

func (a Assign) Eval(env *obj.Env) obj.Object {
	if left, ok := a.l.(Identifier); ok {
		right := obj.Unwrap(a.r.Eval(env))
		if takesPrecedenceNoError(right) {
			return right
		}
		return env.Set(left.String(), right)
	}

	left := a.l.Eval(env)
	if s, ok := left.(obj.Setter); ok {
		right := obj.Unwrap(a.r.Eval(env))
		if takesPrecedence(right) {
			return right
		}
		return s.Set(right)
	}

	return obj.NewError("cannot assign to literal")
}

func (a Assign) String() string {
	return fmt.Sprintf("(%v = %v)", a.l, a.r)
}

func (a Assign) Compile(c *compiler.Compiler) (position int, err error) {
	switch left := a.l.(type) {
	case Identifier:
		symbol := c.Define(string(left))
		if position, err = a.r.Compile(c); err != nil {
			return
		}

		if symbol.Scope == compiler.GlobalScope {
			return c.Emit(code.OpSetGlobal, symbol.Index), nil
		} else {
			return c.Emit(code.OpSetLocal, symbol.Index), nil
		}

	case Dot, Index:
		if position, err = a.l.Compile(c); err != nil {
			return
		}
		if position, err = a.r.Compile(c); err != nil {
			return
		}
		return c.Emit(code.OpDefine), nil

	default:
		return 0, fmt.Errorf("cannot assign to literal")
	}
}

func (a Assign) IsConstExpression() bool {
	return false
}
