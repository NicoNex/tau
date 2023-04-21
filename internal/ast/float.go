package ast

import (
	"strconv"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type Float float64

func NewFloat(f float64) Node {
	return Float(f)
}

func (f Float) Eval() (cobj.Object, error) {
	return cobj.NewFloat(float64(f)), nil
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f Float) Compile(c *compiler.Compiler) (position int, err error) {
	return c.Emit(code.OpConstant, c.AddConstant(cobj.NewFloat(float64(f)))), nil
}

func (f Float) IsConstExpression() bool {
	return true
}
