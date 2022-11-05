package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseNot struct {
	n Node
}

func NewBitwiseNot(n Node) Node {
	return BitwiseNot{n}
}

func (b BitwiseNot) Eval(env *obj.Env) obj.Object {
	var value = obj.Unwrap(b.n.Eval(env))

	if takesPrecedence(value) {
		return value
	}

	if !obj.AssertTypes(value, obj.IntType) {
		return obj.NewError("unsupported operator '~' for type %v", value.Type())
	}

	n := value.(obj.Integer).Val()
	return obj.NewInteger(^n)
}

func (b BitwiseNot) String() string {
	return fmt.Sprintf("~%v", b.n)
}

func (b BitwiseNot) Compile(c *compiler.Compiler) (position int, err error) {
	if b.IsConstExpression() {
		return c.Emit(code.OpConstant, c.AddConstant(b.Eval(nil))), nil
	}

	if position, err = b.n.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpBwNot), nil
}

func (b BitwiseNot) IsConstExpression() bool {
	return b.n.IsConstExpression()
}
