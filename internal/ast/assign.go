package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Assign struct {
	l   Node
	r   Node
	pos int
}

func NewAssign(l, r Node, pos int) Node {
	return Assign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (a Assign) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.Assign: not a constant expression")
}

func (a Assign) String() string {
	return fmt.Sprintf("(%v = %v)", a.l, a.r)
}

func (a Assign) Compile(c *compiler.Compiler) (position int, err error) {
	defer c.Bookmark(position)

	switch left := a.l.(type) {
	case Identifier:
		if position, err = a.r.Compile(c); err != nil {
			return
		}

		symbol, ok := c.Resolve(left.String())
		if !ok {
			symbol = c.Define(left.String())
		}

		if symbol.Scope == compiler.GlobalScope {
			position = c.Emit(code.OpSetGlobal, symbol.Index)
			c.Bookmark(a.pos)
			return
		} else {
			position = c.Emit(code.OpSetLocal, symbol.Index)
			c.Bookmark(a.pos)
			return
		}

	case Dot, Index:
		if position, err = a.l.Compile(c); err != nil {
			return
		}
		if position, err = a.r.Compile(c); err != nil {
			return
		}
		position = c.Emit(code.OpDefine)
		c.Bookmark(a.pos)
		return

	default:
		return 0, fmt.Errorf("cannot assign to literal")
	}
}

func (a Assign) IsConstExpression() bool {
	return false
}
