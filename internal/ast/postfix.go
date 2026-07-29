package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

// PostfixPlusPlus and PostfixMinusMinus are `i++` and `i--`, which increment
// like the prefix ones but evaluate to the value the variable had before, the
// way C does it. The prefix forms are plain `i = i + 1`, and an assignment
// leaves what it assigned behind, so they already evaluate to the new value.
type PostfixPlusPlus struct {
	l   Node
	pos int
}

type PostfixMinusMinus struct {
	l   Node
	pos int
}

func NewPostfixPlusPlus(l Node, pos int) Node {
	return PostfixPlusPlus{l: l, pos: pos}
}

func NewPostfixMinusMinus(l Node, pos int) Node {
	return PostfixMinusMinus{l: l, pos: pos}
}

func (p PostfixPlusPlus) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.PostfixPlusPlus: not a constant expression")
}

func (m PostfixMinusMinus) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.PostfixMinusMinus: not a constant expression")
}

func (p PostfixPlusPlus) String() string {
	return fmt.Sprintf("%v++", p.l)
}

func (m PostfixMinusMinus) String() string {
	return fmt.Sprintf("%v--", m.l)
}

// compilePostfix leaves the old value of l on the stack and increments it by
// n. The old value goes down first, then the assignment runs and puts the new
// one on top of it, and that one is dropped: what is left is what the
// expression is worth.
func compilePostfix(c *compiler.Compiler, l Node, n Node, pos int) (position int, err error) {
	if position, err = l.Compile(c); err != nil {
		return
	}

	// The target is compiled a second time by the assignment, so an index or a
	// field access is evaluated twice, exactly as the prefix forms already do.
	a := Assign{l: l, r: n, pos: pos}
	if position, err = a.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpPop)

	c.Bookmark(pos)
	return
}

func (p PostfixPlusPlus) Compile(c *compiler.Compiler) (position int, err error) {
	return compilePostfix(c, p.l, Plus{l: p.l, r: Integer(1), pos: p.pos}, p.pos)
}

func (m PostfixMinusMinus) Compile(c *compiler.Compiler) (position int, err error) {
	return compilePostfix(c, m.l, Minus{l: m.l, r: Integer(1), pos: m.pos}, m.pos)
}

func (p PostfixPlusPlus) IsConstExpression() bool {
	return false
}

func (m PostfixMinusMinus) IsConstExpression() bool {
	return false
}
