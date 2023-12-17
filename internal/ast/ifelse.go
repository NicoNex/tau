package ast

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type IfExpr struct {
	cond   Node
	body   Node
	altern Node
	pos    int
}

func NewIfExpr(cond, body, alt Node, pos int) Node {
	return IfExpr{
		cond:   cond,
		body:   body,
		altern: alt,
		pos:    pos,
	}
}

func (i IfExpr) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.IfExpr: not a constant expression")
}

func (i IfExpr) String() string {
	if i.altern != nil {
		return fmt.Sprintf("if %v { %v } else { %v }", i.cond, i.body, i.altern)
	}
	return fmt.Sprintf("if %v { %v }", i.cond, i.body)
}

func (i IfExpr) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = i.cond.Compile(c); err != nil {
		return
	}
	jumpNotTruthyPos := c.Emit(code.OpJumpNotTruthy, compiler.GenericPlaceholder)
	if position, err = i.body.Compile(c); err != nil {
		return
	}

	if c.LastIs(code.OpPop) {
		c.RemoveLast()
	}

	jumpPos := c.Emit(code.OpJump, compiler.GenericPlaceholder)
	c.ReplaceOperand(jumpNotTruthyPos, c.Pos())

	if i.altern == nil {
		c.Emit(code.OpNull)
	} else {
		if position, err = i.altern.Compile(c); err != nil {
			return
		}

		if c.LastIs(code.OpPop) {
			c.RemoveLast()
		}
	}

	c.ReplaceOperand(jumpPos, c.Pos())
	return c.Pos(), nil
}

func (i IfExpr) IsConstExpression() bool {
	return false
}
