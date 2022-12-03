package ast

import (
	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Boolean bool

func NewBoolean(b bool) Node {
	return Boolean(b)
}

func (b Boolean) Eval(env *obj.Env) obj.Object {
	return obj.ParseBool(bool(b))
}

func (b Boolean) String() string {
	if b {
		return "true"
	}
	return "false"
}

func (b Boolean) Compile(c *compiler.Compiler) (position int, err error) {
	if bool(b) {
		return c.Emit(code.OpTrue), nil
	} else {
		return c.Emit(code.OpFalse), nil
	}
}

func (b Boolean) IsConstExpression() bool {
	return true
}
