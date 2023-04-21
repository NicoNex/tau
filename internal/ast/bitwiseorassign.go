package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type BitwiseOrAssign struct {
	l   Node
	r   Node
	pos int
}

func NewBitwiseOrAssign(l, r Node, pos int) Node {
	return BitwiseOrAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (b BitwiseOrAssign) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.BitwiseOrAssign: not a constant expression")
}

func (b BitwiseOrAssign) String() string {
	return fmt.Sprintf("(%v |= %v)", b.l, b.r)
}

func (b BitwiseOrAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: b.l, r: BitwiseOr{l: b.l, r: b.r, pos: b.pos}, pos: b.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (b BitwiseOrAssign) IsConstExpression() bool {
	return false
}
