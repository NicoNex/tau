package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Greater struct {
	l Node
	r Node
}

func NewGreater(l, r Node) Node {
	return Greater{l, r}
}

func (g Greater) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(g.l.Eval(env))
		right = obj.Unwrap(g.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '>' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '>' for type %v", right.Type())
	}

	if assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType) {
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return obj.ParseBool(l > r)
	}

	left, right = toFloat(left, right)
	l := left.(*obj.Float).Val()
	r := right.(*obj.Float).Val()
	return obj.ParseBool(l > r)
}

func (g Greater) String() string {
	return fmt.Sprintf("(%v > %v)", g.l, g.r)
}

func (g Greater) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = g.l.Compile(c); err != nil {
		return
	}
	if position, err = g.r.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpGreaterThan), nil
}
