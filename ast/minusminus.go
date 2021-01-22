package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type MinusMinus struct {
	l Node
}

func NewMinusMinus(l Node) Node {
	return MinusMinus{l}
}

func (m MinusMinus) Eval(env *obj.Env) obj.Object {
	var name string
	var left = m.l.Eval(env)

	if isError(left) {
		return left
	}

	if ident, ok := m.l.(Identifier); ok {
		name = ident.String()
	}

	if !assertTypes(left, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '--' for type %v", left.Type())
	}

	if assertTypes(left, obj.INT) {
		l := left.(*obj.Integer).Val()
		return env.Set(name, obj.NewInteger(l-1))
	}

	left, _ = toFloat(left, obj.NullObj)
	l := left.(*obj.Float).Val()
	return env.Set(name, obj.NewFloat(l-1))
}

func (m MinusMinus) String() string {
	return fmt.Sprintf("%v--", m.l)
}
