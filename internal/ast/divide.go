package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Divide struct {
	l   Node
	r   Node
	pos int
}

func NewDivide(l, r Node) Node {
	return Divide{
		l:   l,
		r:   r,
		pos: 0,
	}
}

func (d Divide) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(d.l.Eval(env))
		right = obj.Unwrap(d.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '/' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '/' for type %v", right.Type())
	}

	left, right = toFloat(left, right)
	l := left.(*obj.Float).Val()
	r := right.(*obj.Float).Val()
	return obj.NewFloat(l / r)
}

func (d Divide) String() string {
	return fmt.Sprintf("(%v / %v)", d.l, d.r)
}

func (d Divide) Compile(c *compiler.Compiler) (position int, err error) {
	if d.IsConstExpression() {
		return c.Emit(code.OpConstant, c.AddConstant(d.Eval(nil))), nil
	}

	if position, err = d.l.Compile(c); err != nil {
		return
	}
	if position, err = d.r.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpDiv), nil
}

func (d Divide) IsConstExpression() bool {
	return d.l.IsConstExpression() && d.r.IsConstExpression()
}
