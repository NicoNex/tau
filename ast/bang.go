package ast

import (
	"fmt"
	"github.com/NicoNex/tau/obj"
)

type Bang struct {
	n Node
}

func NewBang(n Node) Node {
	return Bang{n}
}

func (b Bang) Eval(env *obj.Env) obj.Object {
	var value = b.n.Eval(env)

	if isError(value) {
		return value
	}

	switch v := value.(type) {
	case *obj.Boolean:
		return obj.ParseBool(!v.Val())

	case *obj.Null:
		return obj.True

	default:
		return obj.False
	}
}

func (b Bang) String() string {
	return fmt.Sprintf("!%v", b.n)
}
