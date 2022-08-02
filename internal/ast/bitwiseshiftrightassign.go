package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseShiftRightAssign struct {
	l Node
	r Node
}

func NewBitwiseShiftRightAssign(l, r Node) Node {
	return BitwiseShiftRightAssign{l, r}
}

func (b BitwiseShiftRightAssign) Eval(env *obj.Env) obj.Object {
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
		return obj.NewError("unsupported operator '>>=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '>>=' for type %v", right.Type())
	}

	if gs, ok := left.(obj.GetSetter); ok {
		l := gs.Object().(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		return gs.Set(obj.NewInteger(l >> r))
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return env.Set(name, obj.NewInteger(l>>r))
}

func (b BitwiseShiftRightAssign) String() string {
	return fmt.Sprintf("(%v >> %v)", b.l, b.r)
}

func (b BitwiseShiftRightAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{b.l, BitwiseRightShift{b.l, b.r}}
	return n.Compile(c)
}