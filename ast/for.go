package ast

import (
	"fmt"

	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type For struct {
	before Node
	after  Node
	cond   Node
	body   Node
}

func NewFor(cond, body, before, after Node) Node {
	return For{before, after, cond, body}
}

func (f For) Eval(env *obj.Env) obj.Object {
	if f.before != nil {
		obj.Unwrap(f.before.Eval(env))
	}

loop:
	for isTruthy(obj.Unwrap(f.cond.Eval(env))) {
		switch o := obj.Unwrap(f.body.Eval(env)); {
		case o == nil:
			break

		case isError(o) || isReturn(o):
			return o

		case isBreak(o):
			break loop
		}

		if f.after != nil {
			obj.Unwrap(f.after.Eval(env))
		}
	}
	return obj.NullObj
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

	jumpNotTruthyPos := c.Emit(code.OpJumpNotTruthy, 9999)

	if position, err = f.body.Compile(c); err != nil {
		return
	}

	if f.after != nil {
		if position, err = f.after.Compile(c); err != nil {
			return
		}
		c.Emit(code.OpPop)
	}

	c.Emit(code.OpJump, startPos)
	c.ReplaceOperand(jumpNotTruthyPos, c.Pos())
	return c.Pos(), nil
}
