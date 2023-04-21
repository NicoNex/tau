package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type DivideAssign struct {
	l   Node
	r   Node
	pos int
}

func NewDivideAssign(l, r Node, pos int) Node {
	return DivideAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (d DivideAssign) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.DivideAssign: not a constant expression")
}

func (d DivideAssign) String() string {
	return fmt.Sprintf("(%v /= %v)", d.l, d.r)
}

func (d DivideAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: d.l, r: Divide{l: d.l, r: d.r, pos: d.pos}, pos: d.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (d DivideAssign) IsConstExpression() bool {
	return false
}
