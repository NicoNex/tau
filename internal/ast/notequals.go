package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type NotEquals struct {
	l Node
	r Node
}

func NewNotEquals(l, r Node) Node {
	return NotEquals{l, r}
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

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return obj.NewError("unsupported operator '!=' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return obj.NewError("unsupported operator '!=' for type %v", right.Type())
	}

	switch {
	case obj.AssertTypes(right, obj.BoolType, obj.NullType) || obj.AssertTypes(right, obj.BoolType, obj.NullType):
		return obj.ParseBool(left != right)

	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		l := left.(obj.String)
		r := right.(obj.String)
		return obj.ParseBool(l != r)

	case obj.AssertTypes(left, obj.IntType) && obj.AssertTypes(right, obj.IntType):
		l := left.(obj.Integer)
		r := right.(obj.Integer)
		return obj.ParseBool(l != r)

	case obj.AssertTypes(left, obj.FloatType, obj.IntType) && obj.AssertTypes(right, obj.FloatType, obj.IntType):
		left, right = obj.ToFloat(left, right)
		l := left.(obj.Float)
		r := right.(obj.Float)
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
