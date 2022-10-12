package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Mod struct {
	l   Node
	r   Node
	pos int
}

func NewMod(l, r Node, pos int) Node {
	return Mod{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (m Mod) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(m.l.Eval(env))
		right = obj.Unwrap(m.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType) {
		return obj.NewError("unsupported operator '%%' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '%%' for type %v", right.Type())
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()

	if r == 0 {
		return obj.NewError("can't divide by 0")
	}
	return obj.NewInteger(l % r)
}

func (m Mod) String() string {
	return fmt.Sprintf("(%v %% %v)", m.l, m.r)
}

func (m Mod) Compile(c *compiler.Compiler) (position int, err error) {
	if m.IsConstExpression() {
		position = c.Emit(code.OpConstant, c.AddConstant(m.Eval(nil)))
		c.Bookmark(m.pos)
		return
	}

	if position, err = m.l.Compile(c); err != nil {
		return
	}
	if position, err = m.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpMod)
	c.Bookmark(m.pos)
	return
}

func (m Mod) IsConstExpression() bool {
	return m.l.IsConstExpression() && m.r.IsConstExpression()
}
