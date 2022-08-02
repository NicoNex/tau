package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type IfExpr struct {
	cond   Node
	body   Node
	altern Node
}

func NewIfExpr(cond, body, alt Node) Node {
	return IfExpr{cond, body, alt}
}

func (i IfExpr) Eval(env *obj.Env) obj.Object {
	var cond = obj.Unwrap(i.cond.Eval(env))

	if takesPrecedence(cond) {
		return cond
	}

	switch c := cond.(type) {
	case *obj.Boolean:
		if c.Val() {
			return obj.Unwrap(i.body.Eval(env))
		}
		return i.alternative(env)

	case *obj.Null:
		return i.alternative(env)

	default:
		return obj.Unwrap(i.body.Eval(env))
	}
}

func (i IfExpr) String() string {
	if i.altern != nil {
		return fmt.Sprintf("if %v { %v } else { %v }", i.cond, i.body, i.altern)
	}
	return fmt.Sprintf("if %v { %v }", i.cond, i.body)
}

func (i IfExpr) alternative(env *obj.Env) obj.Object {
	if i.altern != nil {
		return obj.Unwrap(i.altern.Eval(env))
	}
	return obj.NullObj
}

func (i IfExpr) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = i.cond.Compile(c); err != nil {
		return
	}
	jumpNotTruthyPos := c.Emit(code.OpJumpNotTruthy, 9999)
	if position, err = i.body.Compile(c); err != nil {
		return
	}

	if c.LastIs(code.OpPop) {
		c.RemoveLast()
	}

	jumpPos := c.Emit(code.OpJump, 9999)
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
