package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseAndAssign struct {
	l Node
	r Node
}

func NewBitwiseAndAssign(l, r Node) Node {
	return BitwiseAndAssign{l, r}
}

func (b BitwiseAndAssign) Eval(env *obj.Env) obj.Object {
	var (
		name  string
		left  = b.l.Eval(env)
		right = obj.Unwrap(b.r.Eval(env))
	)

	if ident, ok := b.l.(Identifier); ok {
		name = ident.String()
	}

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType) {
		return obj.NewError("unsupported operator '&=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '&=' for type %v", right.Type())
	}

	if gs, ok := left.(obj.GetSetter); ok {
		l := gs.Object().(obj.Integer).Val()
		r := right.(obj.Integer).Val()
		return gs.Set(obj.NewInteger(l & r))
	}

	l := left.(obj.Integer).Val()
	r := right.(obj.Integer).Val()
	return env.Set(name, obj.NewInteger(l&r))
}

func (b BitwiseAndAssign) String() string {
	return fmt.Sprintf("(%v &= %v)", b.l, b.r)
}

func (b BitwiseAndAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{b.l, BitwiseAnd{b.l, b.r}}
	return n.Compile(c)
}

func (b BitwiseAndAssign) IsConstExpression() bool {
	return false
}
