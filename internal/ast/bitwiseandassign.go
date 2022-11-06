package ast

import (
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

func (b BitwiseAndAssign) Eval(env *obj.Env) obj.Object {
	var (
		name  string
		left  = b.l.Eval(env)
		right = obj.Unwrap(b.r.Eval(env))
	)

	if ident, ok := b.l.(Identifier); ok {
		name = ident.String()
	}

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !obj.AssertTypes(left, obj.IntType) {
		return obj.NewError("unsupported operator '&=' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '&=' for type %v", right.Type())
	}

	if gs, ok := left.(obj.GetSetter); ok {
		l := gs.Object().(obj.Integer)
		r := right.(obj.Integer)
		return gs.Set(obj.Integer(l & r))
	}

	l := left.(obj.Integer)
	r := right.(obj.Integer)
	return env.Set(name, obj.Integer(l&r))
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
