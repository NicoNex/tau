package ast

import (
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

func (i Identifier) Eval(env *obj.Env) obj.Object {
	if c, ok := env.Get(i.name); ok {
		return c
	} else if o, ok := obj.ResolveBuiltin(i.name); ok {
		return o
	}

	return obj.NewError("name %q is not defined", i)
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
