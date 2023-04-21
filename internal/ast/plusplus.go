package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type PlusPlus struct {
	r   Node
	pos int
}

func NewPlusPlus(r Node, pos int) Node {
	return PlusPlus{
		r:   r,
		pos: pos,
	}
}

func (p PlusPlus) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.PlusPlus: not a constant expression")
}

func (p PlusPlus) String() string {
	return fmt.Sprintf("++%v", p.r)
}

func (p PlusPlus) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: p.r, r: Plus{l: p.r, r: Integer(1), pos: p.pos}, pos: p.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (p PlusPlus) IsConstExpression() bool {
	return false
}
