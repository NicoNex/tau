package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Bang struct {
	n Node
}

func NewBang(n Node) Node {
	return Bang{n}
}

func (b Bang) Eval(env *obj.Env) obj.Object {
	var value = obj.Unwrap(b.n.Eval(env))

	if takesPrecedence(value) {
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

func (b Bang) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = b.n.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpBang), nil
}
