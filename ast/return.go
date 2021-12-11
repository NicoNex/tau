package ast

import (
	"fmt"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type Return struct {
	v Node
}

func NewReturn(n Node) Node {
	return Return{n}
}

func (r Return) Eval(env *obj.Env) obj.Object {
	return obj.NewReturn(r.v.Eval(env))
}

func (r Return) String() string {
	return fmt.Sprintf("return %v", r.v)
}

func (r Return) Compile(c *compiler.Compiler) (position int) {
	return 0
}
