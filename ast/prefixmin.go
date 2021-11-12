package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type PrefixMinus struct {
	n Node
}

func NewPrefixMinus(n Node) Node {
	return PrefixMinus{n}
}

func (p PrefixMinus) Eval(env *obj.Env) obj.Object {
	var value = unwrap(p.n.Eval(env))

	if takesPrecedence(value) {
		return value
	}

	switch v := value.(type) {
	case *obj.Integer:
		return obj.NewInteger(-v.Val())

	case *obj.Float:
		return obj.NewFloat(-v.Val())

	default:
		return obj.NewError("unsupported prefix operator '-' for type %v", value.Type())

	}
}

func (p PrefixMinus) String() string {
	return fmt.Sprintf("-%v", p.n)
}
