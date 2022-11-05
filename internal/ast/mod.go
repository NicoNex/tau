package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Mod struct {
	l Node
	r Node
}

func NewMod(l, r Node) Node {
	return Mod{l, r}
}

func (m Mod) Eval(env *obj.Env) obj.Object {
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

	if !obj.AssertTypes(left, obj.IntType) {
		return obj.NewError("unsupported operator '%%' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '%%' for type %v", right.Type())
	}

	l := left.(obj.Integer).Val()
	r := right.(obj.Integer).Val()

	if r == 0 {
		return obj.NewError("can't divide by 0")
	}
	return obj.NewInteger(l % r)
}

func (m Mod) String() string {
	return fmt.Sprintf("(%v %% %v)", m.l, m.r)
}

func (m Mod) Compile(c *compiler.Compiler) (position int, err error) {
	if m.IsConstExpression() {
		return c.Emit(code.OpConstant, c.AddConstant(m.Eval(nil))), nil
	}

	if position, err = m.l.Compile(c); err != nil {
		return
	}
	if position, err = m.r.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpMod), nil
}

func (m Mod) IsConstExpression() bool {
	return m.l.IsConstExpression() && m.r.IsConstExpression()
}
