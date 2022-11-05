package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type PlusPlus struct {
	r Node
}

func NewPlusPlus(r Node) Node {
	return PlusPlus{r}
}

func (p PlusPlus) Eval(env *obj.Env) obj.Object {
	var (
		name  string
		right = p.r.Eval(env)
	)

	if takesPrecedence(right) {
		return right
	}

	if ident, ok := p.r.(Identifier); ok {
		name = ident.String()
	}

	if !obj.AssertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '++' for type %v", right.Type())
	}

	if obj.AssertTypes(right, obj.IntType) {
		if gs, ok := right.(obj.GetSetter); ok {
			r := gs.Object().(obj.Integer)
			return gs.Set(obj.Integer(r + 1))
		}

		r := right.(obj.Integer)
		return env.Set(name, obj.Integer(r+1))
	}

	if gs, ok := right.(obj.GetSetter); ok {
		rightFl, _ := obj.ToFloat(gs.Object(), obj.NullObj)
		r := rightFl.(obj.Float)
		return gs.Set(obj.Float(r + 1))
	}

	rightFl, _ := obj.ToFloat(right, obj.NullObj)
	r := rightFl.(obj.Float)
	return env.Set(name, obj.Float(r+1))
}

func (p PlusPlus) String() string {
	return fmt.Sprintf("++%v", p.r)
}

func (p PlusPlus) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{p.r, Plus{p.r, Integer(1)}}
	return n.Compile(c)
}

func (p PlusPlus) IsConstExpression() bool {
	return false
}
