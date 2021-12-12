package ast

import (
	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
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

func (b Boolean) Compile(c *compiler.Compiler) (position int) {
	isTrue := bool(b)

	if isTrue {
		return c.Emit(code.OpTrue)
	} else {
		return c.Emit(code.OpFalse)
	}
}
