package ast

import (
	"errors"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Identifier struct {
	name string
	pos  int
}

func NewIdentifier(name string, pos int) Identifier {
	return Identifier{
		name: name,
		pos:  pos,
	}
}

func (i Identifier) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.Identifier: not a constant expression")
}

func (i Identifier) String() string {
	return i.name
}

func (i Identifier) Compile(c *compiler.Compiler) (position int, err error) {
	if symbol, ok := c.Resolve(i.name); ok {
		return c.LoadSymbol(symbol), nil
	}

	// The name may be defined further down the file: reserve a global for it
	// and check at the end of the compilation that it showed up.
	return c.LoadSymbol(c.DefineForward(i.name, i.pos)), nil
}

func (i Identifier) IsConstExpression() bool {
	return false
}
