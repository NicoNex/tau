package ast

import (
	"fmt"
	"tau/obj"
)

type Bang struct {
	n Node
}

func NewBang(n Node) Node {
	return Bang{n}
}

func (b Bang) Eval() obj.Object {
	var value = b.n.Eval()

	if isError(value) {
		return value
	}

	switch v := value.(type) {
	case *obj.Boolean:
		return obj.ParseBool(!v.Val())

	case *obj.Null:
		return obj.False

	default:
		return obj.True
	}
}

func (b Bang) String() string {
	return fmt.Sprintf("-%v", b.n)
}
