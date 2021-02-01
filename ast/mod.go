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
	var left = m.l.Eval(env)
	var right = m.r.Eval(env)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT) {
		return obj.NewError("unsupported operator '%%' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT) {
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
