package ast

import (
	"fmt"
	"tau/obj"
)

type PrefixMinus struct {
	n Node
}

func NewPrefixMinus(n Node) Node {
	return PrefixMinus{n}
}

func (p PrefixMinus) Eval() obj.Object {
	var value = p.n.Eval()

	if isError(value) {
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
