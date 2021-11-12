package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
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
		left  = unwrap(m.l.Eval(env))
		right = unwrap(m.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType) {
		return obj.NewError("unsupported operator '%%' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '%%' for type %v", right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()

	if r == 0 {
		return obj.NewError("can't divide by 0")
	}
	return obj.NewInteger(l % r)
}

func (m Mod) String() string {
	return fmt.Sprintf("(%v %% %v)", m.l, m.r)
}
