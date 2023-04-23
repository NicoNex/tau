package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type MinusMinus struct {
	r   Node
	pos int
}

func NewMinusMinus(r Node, pos int) Node {
	return MinusMinus{
		r:   r,
		pos: pos,
	}
}

func (m MinusMinus) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.MinusMinus: not a constant expression")
}

func (m MinusMinus) String() string {
	return fmt.Sprintf("--%v", m.r)
}

func (m MinusMinus) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: m.r, r: Minus{l: m.r, r: Integer(1), pos: m.pos}, pos: m.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (m MinusMinus) IsConstExpression() bool {
	return false
}
