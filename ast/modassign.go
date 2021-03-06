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
	var name string
	var left = m.l.Eval(env)
	var right = m.r.Eval(env)

	if ident, ok := m.l.(Identifier); ok {
		name = ident.String()
	} else {
		return obj.NewError("cannot assign to literal")
	}

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT) {
		return obj.NewError("unsupported operator '%%=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT) {
		return obj.NewError("unsupported operator '%%=' for type %v", right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()

	if r == 0 {
		return obj.NewError("can't divide by 0")
	}
	return env.Set(name, obj.NewInteger(l%r))
}

func (m ModAssign) String() string {
	return fmt.Sprintf("(%v %%= %v)", m.l, m.r)
}
