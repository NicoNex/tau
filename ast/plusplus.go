package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type PlusPlus struct {
	l Node
}

func NewPlusPlus(l Node) Node {
	return PlusPlus{l}
}

func (m PlusPlus) Eval(env *obj.Env) obj.Object {
	var left = m.l.Eval(env)
	var name string

	if isError(left) {
		return left
	}

	if ident, ok := m.l.(Identifier); ok {
		name = ident.String()
	}

	if !assertTypes(left, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '++' for type %v", left.Type())
	}

	if assertTypes(left, obj.INT) {
		l := left.(*obj.Integer).Val()
		env.Set(name, obj.NewInteger(l+1))
		return obj.NullObj
	}

	left, _ = toFloat(left, obj.NullObj)
	l := left.(*obj.Float).Val()
	env.Set(name, obj.NewFloat(l+1))

	return obj.NullObj
}

func (m PlusPlus) String() string {
	return fmt.Sprintf("%v++", m.l)
}
