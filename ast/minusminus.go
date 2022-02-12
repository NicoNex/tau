package ast

import (
	"fmt"

	"github.com/NicoNex/tau/compiler"
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
		name  string
		right = m.r.Eval(env)
	)

	if takesPrecedence(right) {
		return right
	}

	if ident, ok := m.r.(Identifier); ok {
		name = ident.String()
	}

	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '--' for type %v", right.Type())
	}

	if assertTypes(right, obj.IntType) {
		if gs, ok := right.(obj.GetSetter); ok {
			r := gs.Object().(*obj.Integer).Val()
			return gs.Set(obj.NewInteger(r - 1))
		}

		r := right.(*obj.Integer).Val()
		return env.Set(name, obj.NewInteger(r-1))
	}

	if gs, ok := right.(obj.GetSetter); ok {
		rightFl, _ := toFloat(gs.Object(), obj.NullObj)
		r := rightFl.(*obj.Float).Val()
		return gs.Set(obj.NewFloat(r - 1))
	}

	rightFl, _ := toFloat(right, obj.NullObj)
	r := rightFl.(*obj.Float).Val()
	return env.Set(name, obj.NewFloat(r-1))
}

func (m MinusMinus) String() string {
	return fmt.Sprintf("--%v", m.r)
}

func (m MinusMinus) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{m.r, Minus{m.r, Integer(1)}}
	return n.Compile(c)
}
