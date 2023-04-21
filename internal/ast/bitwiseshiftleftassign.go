package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type BitwiseShiftLeftAssign struct {
	l   Node
	r   Node
	pos int
}

func NewBitwiseShiftLeftAssign(l, r Node, pos int) Node {
	return BitwiseShiftLeftAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (b BitwiseShiftLeftAssign) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.BitwiseShiftLeftAssign: not a constant expression")
}

func (b BitwiseShiftLeftAssign) String() string {
	return fmt.Sprintf("(%v << %v)", b.l, b.r)
}

func (b BitwiseShiftLeftAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: b.l, r: BitwiseLeftShift{l: b.l, r: b.r, pos: b.pos}, pos: b.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (b BitwiseShiftLeftAssign) IsConstExpression() bool {
	return false
}
