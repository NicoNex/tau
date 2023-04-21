package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type Index struct {
	left  Node
	index Node
	pos   int
}

func NewIndex(l, i Node, pos int) Node {
	return Index{left: l,
		index: i,
		pos:   pos,
	}
}

func (i Index) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.Index: not a constant expression")
}

func (i Index) String() string {
	return fmt.Sprintf("%v[%v]", i.left, i.index)
}

func (i Index) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = i.left.Compile(c); err != nil {
		return
	}
	if position, err = i.index.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpIndex)
	c.Bookmark(i.pos)
	return
}

func (i Index) IsConstExpression() bool {
	return false
}
