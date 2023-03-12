package ast

import (
	"strconv"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Integer int64

func NewInteger(i int64) Node {
	return Integer(i)
}

func (i Integer) Eval(env *obj.Env) obj.Object {
	return obj.Integer(int64(i))
}

func (i Integer) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i Integer) Compile(c *compiler.Compiler) (position int, err error) {
	return c.Emit(code.OpConstant, c.AddConstant(c.NewInteger(int64(i)))), nil
}

func (i Integer) IsConstExpression() bool {
	return true
}
