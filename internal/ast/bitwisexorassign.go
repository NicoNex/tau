package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type BitwiseXorAssign struct {
	l   Node
	r   Node
	pos int
}

func NewBitwiseXorAssign(l, r Node, pos int) Node {
	return BitwiseXorAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (b BitwiseXorAssign) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.BitwiseXorAssign: not a constant expression")
}

func (b BitwiseXorAssign) String() string {
	return fmt.Sprintf("(%v ^= %v)", b.l, b.r)
}

func (b BitwiseXorAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: b.l, r: BitwiseXor{l: b.l, r: b.r, pos: b.pos}, pos: b.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (b BitwiseXorAssign) IsConstExpression() bool {
	return false
}
