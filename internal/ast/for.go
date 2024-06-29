package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type For struct {
	before Node
	after  Node
	cond   Node
	body   Node
	pos    int
}

func NewFor(cond, body, before, after Node, pos int) Node {
	return For{
		before: before,
		after:  after,
		cond:   cond,
		body:   body,
		pos:    pos,
	}
}

func (f For) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.For: not a constant expression")
}

func (f For) String() string {
	return fmt.Sprintf("for %v { %v }", f.cond, f.body)
}

func (f For) Compile(c *compiler.Compiler) (position int, err error) {
	if f.before != nil {
		if position, err = f.before.Compile(c); err != nil {
			return
		}
	}

	startPos := c.Pos()
	if position, err = f.cond.Compile(c); err != nil {
		return
	}

	jumpNotTruthyPos := c.Emit(code.OpJumpNotTruthy, compiler.GenericPlaceholder)

	startBody := c.Pos()
	if position, err = f.body.Compile(c); err != nil {
		return
	}
	endBody := c.Pos()

	if f.after != nil {
		if position, err = f.after.Compile(c); err != nil {
			return
		}
		c.Emit(code.OpPop)
	}

	c.Emit(code.OpJump, startPos)
	endPos := c.Emit(code.OpNull)
	c.ReplaceOperand(jumpNotTruthyPos, endPos)

	err = c.ReplaceContinueOperands(startBody, endBody, endBody)
	if err != nil {
		return
	}
	err = c.ReplaceBreakOperands(startBody, endBody, endPos)
	if err != nil {
		return
	}

	c.Bookmark(f.pos)
	return endPos, nil
}

func (f For) IsConstExpression() bool {
	return false
}
