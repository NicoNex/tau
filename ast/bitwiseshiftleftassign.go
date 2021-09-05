package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type BitwiseShiftLeftAssign struct {
	l Node
	r Node
}

func NewBitwiseShiftLeftAssign(l, r Node) Node {
	return BitwiseShiftLeftAssign{l, r}
}

func (b BitwiseShiftLeftAssign) Eval(env *obj.Env) obj.Object {
	var (
		name        string
		isContainer bool
		left        = b.l.Eval(env)
		right       = unwrap(b.r.Eval(env))
	)

	if ident, ok := b.l.(Identifier); ok {
		name = ident.String()
	} else if _, isContainer = left.(*obj.Container); !isContainer {
		return obj.NewError("cannot assign to literal")
	}

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT) {
		return obj.NewError("unsupported operator '<<=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT) {
		return obj.NewError("unsupported operator '<<=' for type %v", right.Type())
	}
	l := unwrap(left).(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()

	if isContainer {
		return left.(*obj.Container).Set(obj.NewInteger(l << r))
	}
	return env.Set(name, obj.NewInteger(l<<r))
}

func (b BitwiseShiftLeftAssign) String() string {
	return fmt.Sprintf("(%v << %v)", b.l, b.r)
}
