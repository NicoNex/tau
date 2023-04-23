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
	return 0, c.UnresolvedError(i.name, i.pos)
}

func (i Identifier) IsConstExpression() bool {
	return false
}
