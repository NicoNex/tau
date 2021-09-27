package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type MinusMinus struct {
	r Node
}

func NewMinusMinus(r Node) Node {
	return MinusMinus{r}
}

func (m MinusMinus) Eval(env *obj.Env) obj.Object {
	var (
		name        string
		isContainer bool
		right       = m.r.Eval(env)
	)

	if isError(right) {
		return right
	}

	if ident, ok := m.r.(Identifier); ok {
		name = ident.String()
	} else {
		_, isContainer = right.(*obj.Container)
	}

	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '--' for type %v", right.Type())
	}

	if assertTypes(right, obj.IntType) {
		r := unwrap(right).(*obj.Integer).Val()

		if isContainer {
			return right.(*obj.Container).Set(obj.NewInteger(r - 1))
		}
		return env.Set(name, obj.NewInteger(r-1))
	}

	rightFl, _ := toFloat(unwrap(right), obj.NullObj)
	r := rightFl.(*obj.Float).Val()

	if isContainer {
		return right.(*obj.Container).Set(obj.NewFloat(r - 1))
	}
	return env.Set(name, obj.NewFloat(r-1))
}

func (m MinusMinus) String() string {
	return fmt.Sprintf("--%v", m.r)
}
