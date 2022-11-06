package ast

import (
	"strconv"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Float float64

func NewFloat(f float64) Node {
	return Float(f)
}

func (f Float) Eval(env *obj.Env) obj.Object {
	return obj.Float(float64(f))
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f Float) Compile(c *compiler.Compiler) (position int, err error) {
	return c.Emit(code.OpConstant, c.AddConstant(obj.Float(float64(f)))), nil
}

func (f Float) IsConstExpression() bool {
	return true
}
