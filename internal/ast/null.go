package ast

import (
	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Null struct{}

func NewNull() Null {
	return Null{}
}

func (n Null) Eval(_ *obj.Env) obj.Object {
	return obj.NullObj
}

func (n Null) String() string {
	return "null"
}

func (n Null) Compile(c *compiler.Compiler) (position int, err error) {
	return c.Emit(code.OpConstant, c.AddConstant(obj.NullObj)), nil
}
