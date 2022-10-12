package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Return struct {
	v   Node
	pos int
}

func NewReturn(n Node, pos int) Node {
	return Return{
		v:   n,
		pos: pos,
	}
}

func (r Return) Eval(env *obj.Env) obj.Object {
	return obj.NewReturn(obj.Unwrap(r.v.Eval(env)))
}

func (r Return) String() string {
	return fmt.Sprintf("return %v", r.v)
}

func (r Return) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = r.v.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpReturnValue)
	c.Bookmark(r.pos)
	return
}

func (r Return) IsConstExpression() bool {
	return false
}
