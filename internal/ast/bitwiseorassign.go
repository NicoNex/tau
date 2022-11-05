package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseOrAssign struct {
	l Node
	r Node
}

func NewBitwiseOrAssign(l, r Node) Node {
	return BitwiseOrAssign{l, r}
}

func (b BitwiseOrAssign) Eval(env *obj.Env) obj.Object {
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

	if !obj.AssertTypes(left, obj.IntType) {
		return obj.NewError("unsupported operator '|=' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '|=' for type %v", right.Type())
	}

	if gs, ok := left.(obj.GetSetter); ok {
		l := gs.Object().(obj.Integer)
		r := right.(obj.Integer)
		return gs.Set(obj.Integer(l | r))
	}

	l := left.(obj.Integer)
	r := right.(obj.Integer)
	return env.Set(name, obj.Integer(l|r))
}

func (b BitwiseOrAssign) String() string {
	return fmt.Sprintf("(%v |= %v)", b.l, b.r)
}

func (b BitwiseOrAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{b.l, BitwiseOr{b.l, b.r}}
	return n.Compile(c)
}

func (b BitwiseOrAssign) IsConstExpression() bool {
	return false
}
