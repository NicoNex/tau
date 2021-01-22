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

func (p Mod) Eval(env *obj.Env) obj.Object {
	var left = p.l.Eval(env)
	var right = p.r.Eval(env)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT) {
		return obj.NewError("unsupported operator '%' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT) {
		return obj.NewError("unsupported operator '%' for type %v", right.Type())
	}
	if right.(*obj.Integer).Val() == 0 {
		return obj.NewError("Can't divide by 0")
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()
	return obj.NewInteger(l % r)
}

func (p Mod) String() string {
	return fmt.Sprintf("(%v %% %v)", p.l, p.r)
}
