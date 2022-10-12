package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Identifier string

func NewIdentifier(name string, pos int) Node {
	return Identifier(name)
}

func (i Identifier) Eval(env *obj.Env) obj.Object {
	if c, ok := env.Get(string(i)); ok {
		return c
	} else if o, ok := obj.ResolveBuiltin(string(i)); ok {
		return o
	}

	return obj.NewError("name %q is not defined", i)
}

func (i Identifier) String() string {
	return string(i)
}

func (i Identifier) Compile(c *compiler.Compiler) (position int, err error) {
	if symbol, ok := c.Resolve(string(i)); ok {
		return c.LoadSymbol(symbol), nil
	}
	return 0, fmt.Errorf("undefined variable %s", string(i))
}

func (i Identifier) IsConstExpression() bool {
	return false
}
