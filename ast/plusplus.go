package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type PlusPlus struct {
	r Node
}

func NewPlusPlus(r Node) Node {
	return PlusPlus{r}
}

func (p PlusPlus) Eval(env *obj.Env) obj.Object {
	var right = p.r.Eval(env)
	var name string

	if isError(right) {
		return right
	}

	if ident, ok := p.r.(Identifier); ok {
		name = ident.String()
	}

	if !assertTypes(right, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '++' for type %v", right.Type())
	}

	if assertTypes(right, obj.INT) {
		r := right.(*obj.Integer).Val()
		return env.Set(name, obj.NewInteger(r+1))
	}

	right, _ = toFloat(right, obj.NullObj)
	r := right.(*obj.Float).Val()
	return env.Set(name, obj.NewFloat(r+1))
}

func (p PlusPlus) String() string {
	return fmt.Sprintf("++%v", p.r)
}
