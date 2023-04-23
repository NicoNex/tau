package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseAndAssign struct {
	l   Node
	r   Node
	pos int
}

func NewBitwiseAndAssign(l, r Node, pos int) Node {
	return BitwiseAndAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (b BitwiseAndAssign) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.BitwiseAndAssign: not a constant expression")
}

func (b BitwiseAndAssign) String() string {
	return fmt.Sprintf("(%v &= %v)", b.l, b.r)
}

func (b BitwiseAndAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: b.l, r: BitwiseAnd{l: b.l, r: b.r, pos: b.pos}, pos: b.pos}
	position, err = n.Compile(c)
	c.Bookmark(b.pos)
	return
}

func (b BitwiseAndAssign) IsConstExpression() bool {
	return false
}
