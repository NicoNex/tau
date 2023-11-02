package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type PlusAssign struct {
	l   Node
	r   Node
	pos int
}

func NewPlusAssign(l, r Node, pos int) Node {
	return PlusAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (p PlusAssign) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.PlusAssign: not a constant expression")
}

func (p PlusAssign) String() string {
	return fmt.Sprintf("(%v += %v)", p.l, p.r)
}

func (p PlusAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: p.l, r: Plus{l: p.l, r: p.r, pos: p.pos}, pos: p.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (p PlusAssign) IsConstExpression() bool {
	return false
}
