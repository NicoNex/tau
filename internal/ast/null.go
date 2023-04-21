package ast

import (
	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type Null struct{}

func NewNull() Null {
	return Null{}
}

func (n Null) Eval() (cobj.Object, error) {
	return cobj.NullObj, nil
}

func (n Null) String() string {
	return "null"
}

func (n Null) Compile(c *compiler.Compiler) (position int, err error) {
	return c.Emit(code.OpConstant, c.AddConstant(cobj.NullObj)), nil
}

func (n Null) IsConstExpression() bool {
	return true
}
