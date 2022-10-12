package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Times struct {
	l   Node
	r   Node
	pos int
}

func NewTimes(l, r Node, pos int) Node {
	return Times{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (t Times) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(t.l.Eval(env))
		right = obj.Unwrap(t.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '*' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '*' for type %v", right.Type())
	}

	if assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType) {
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return obj.NewInteger(l * r)
	}

	left, right = toFloat(left, right)
	l := left.(*obj.Float).Val()
	r := right.(*obj.Float).Val()
	return obj.NewFloat(l * r)
}

func (t Times) String() string {
	return fmt.Sprintf("(%v * %v)", t.l, t.r)
}

func (t Times) Compile(c *compiler.Compiler) (position int, err error) {
	if t.IsConstExpression() {
		return c.Emit(code.OpConstant, c.AddConstant(t.Eval(nil))), nil
	}

	if position, err = t.l.Compile(c); err != nil {
		return
	}
	if position, err = t.r.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpMul), nil
}

func (t Times) IsConstExpression() bool {
	return t.l.IsConstExpression() && t.r.IsConstExpression()
}
