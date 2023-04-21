package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type ModAssign struct {
	l   Node
	r   Node
	pos int
}

func NewModAssign(l, r Node, pos int) Node {
	return ModAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (m ModAssign) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.ModAssign: not a constant expression")
}

func (m ModAssign) String() string {
	return fmt.Sprintf("(%v %%= %v)", m.l, m.r)
}

func (m ModAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: m.l, r: Mod{l: m.l, r: m.r, pos: m.pos}, pos: m.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (m ModAssign) IsConstExpression() bool {
	return false
}
