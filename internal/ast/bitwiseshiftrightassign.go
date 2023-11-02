package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseShiftRightAssign struct {
	l   Node
	r   Node
	pos int
}

func NewBitwiseShiftRightAssign(l, r Node, pos int) Node {
	return BitwiseShiftRightAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (b BitwiseShiftRightAssign) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.BitwiseShiftRightAssign: not a constant expression")
}

func (b BitwiseShiftRightAssign) String() string {
	return fmt.Sprintf("(%v >> %v)", b.l, b.r)
}

func (b BitwiseShiftRightAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: b.l, r: BitwiseRightShift{l: b.l, r: b.r, pos: b.pos}, pos: b.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (b BitwiseShiftRightAssign) IsConstExpression() bool {
	return false
}
