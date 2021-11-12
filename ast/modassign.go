package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type ModAssign struct {
	l Node
	r Node
}

func NewModAssign(l, r Node) Node {
	return ModAssign{l, r}
}

func (m ModAssign) Eval(env *obj.Env) obj.Object {
	var (
		name        string
		isContainer bool
		left        = m.l.Eval(env)
		right       = unwrap(m.r.Eval(env))
	)

	if ident, ok := m.l.(Identifier); ok {
		name = ident.String()
	} else if _, isContainer = left.(*obj.Container); !isContainer {
		return obj.NewError("cannot assign to literal")
	}

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType) {
		return obj.NewError("unsupported operator '%%=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '%%=' for type %v", right.Type())
	}

	l := unwrap(left).(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()

	if r == 0 {
		return obj.NewError("can't divide by 0")
	}

	if isContainer {
		return left.(*obj.Container).Set(obj.NewInteger(l % r))
	}
	return env.Set(name, obj.NewInteger(l%r))
}

func (m ModAssign) String() string {
	return fmt.Sprintf("(%v %%= %v)", m.l, m.r)
}
