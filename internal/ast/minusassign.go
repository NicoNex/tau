package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type MinusAssign struct {
	l   Node
	r   Node
	pos int
}

func NewMinusAssign(l, r Node, pos int) Node {
	return MinusAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (m MinusAssign) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.MinusAssign: not a constant expression")
}

func (m MinusAssign) String() string {
	return fmt.Sprintf("(%v -= %v)", m.l, m.r)
}

func (m MinusAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: m.l, r: Minus{l: m.l, r: m.r, pos: m.pos}, pos: m.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (m MinusAssign) IsConstExpression() bool {
	return false
}
