package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
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
		name  string
		left  = m.l.Eval(env)
		right = obj.Unwrap(m.r.Eval(env))
	)

	if ident, ok := m.l.(Identifier); ok {
		name = ident.String()
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

	if gs, ok := left.(obj.GetSetter); ok {
		l := gs.Object().(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		if r == 0 {
			return obj.NewError("can't divide by 0")
		}
		return gs.Set(obj.NewInteger(l % r))
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

func (m ModAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{m.l, Mod{m.l, m.r}}
	return n.Compile(c)
}
