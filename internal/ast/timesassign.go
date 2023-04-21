package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type TimesAssign struct {
	l   Node
	r   Node
	pos int
}

func NewTimesAssign(l, r Node, pos int) Node {
	return TimesAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (t TimesAssign) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.TimesAssign: not a constant expression")
}

func (t TimesAssign) String() string {
	return fmt.Sprintf("(%v *= %v)", t.l, t.r)
}

func (t TimesAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: t.l, r: Times{l: t.l, r: t.r, pos: t.pos}, pos: t.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (t TimesAssign) IsConstExpression() bool {
	return false
}
