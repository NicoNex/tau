package ast

import (
	"strconv"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type String string

func NewString(s string) Node {
	return String(s)
}

func (s String) Eval(env *obj.Env) obj.Object {
	return obj.NewString(string(s))
}

func (s String) String() string {
	return string(s)
}

func (s String) Quoted() string {
	return strconv.Quote(string(s))
}

func (s String) Compile(c *compiler.Compiler) (position int, err error) {
	return c.Emit(code.OpConstant, c.AddConstant(obj.NewString(string(s)))), nil
}
