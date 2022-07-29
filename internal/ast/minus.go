package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Minus struct {
	l Node
	r Node
}

func NewMinus(l, r Node) Node {
	return Minus{l, r}
}

func (m Minus) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(m.l.Eval(env))
		right = obj.Unwrap(m.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '-' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '-' for type %v", right.Type())
	}

	if assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType) {
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return obj.NewInteger(l - r)
	}

	left, right = toFloat(left, right)
	l := left.(*obj.Float).Val()
	r := right.(*obj.Float).Val()
	return obj.NewFloat(l - r)
}

func (m Minus) String() string {
	return fmt.Sprintf("(%v - %v)", m.l, m.r)
}

func (m Minus) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = m.l.Compile(c); err != nil {
		return
	}
	if position, err = m.r.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpSub), nil
}
