package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type NotEquals struct {
	l   Node
	r   Node
	pos int
}

func NewNotEquals(l, r Node, pos int) Node {
	return NotEquals{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (n NotEquals) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(n.l.Eval(env))
		right = obj.Unwrap(n.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return obj.NewError("unsupported operator '!=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return obj.NewError("unsupported operator '!=' for type %v", right.Type())
	}

	switch {
	case assertTypes(right, obj.BoolType, obj.NullType) || assertTypes(right, obj.BoolType, obj.NullType):
		return obj.ParseBool(left != right)

	case assertTypes(left, obj.StringType) && assertTypes(right, obj.StringType):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		return obj.ParseBool(l != r)

	case assertTypes(left, obj.IntType) && assertTypes(right, obj.IntType):
		l := left.(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return obj.ParseBool(l != r)

	case assertTypes(left, obj.FloatType, obj.IntType) && assertTypes(right, obj.FloatType, obj.IntType):
		left, right = toFloat(left, right)
		l := left.(*obj.Float).Val()
		r := right.(*obj.Float).Val()
		return obj.ParseBool(l != r)

	default:
		return obj.True
	}
}

func (n NotEquals) String() string {
	return fmt.Sprintf("(%v != %v)", n.l, n.r)
}

func (n NotEquals) Compile(c *compiler.Compiler) (position int, err error) {
	if n.IsConstExpression() {
		return c.Emit(code.OpConstant, c.AddConstant(n.Eval(nil))), nil
	}

	if position, err = n.l.Compile(c); err != nil {
		return
	}
	if position, err = n.r.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpNotEqual), nil
}

func (n NotEquals) IsConstExpression() bool {
	return n.l.IsConstExpression() && n.r.IsConstExpression()
}
