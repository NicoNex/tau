package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type ModAssign struct {
	l   Node
	r   Node
	pos int
}

func NewModAssign(l, r Node, pos int) Node {
	return ModAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (m ModAssign) Eval(env *obj.Env) obj.Object {
	var (
		name  string
		left  = m.l.Eval(env)
		right = obj.Unwrap(m.r.Eval(env))
	)

	if ident, ok := m.l.(Identifier); ok {
		name = ident.String()
	}

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !assertTypes(left, obj.IntType) {
		return obj.NewError("unsupported operator '%%=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType) {
		return obj.NewError("unsupported operator '%%=' for type %v", right.Type())
	}

	if gs, ok := left.(obj.GetSetter); ok {
		l := gs.Object().(*obj.Integer).Val()
		r := right.(*obj.Integer).Val()
		if r == 0 {
			return obj.NewError("can't divide by 0")
		}
		return gs.Set(obj.NewInteger(l % r))
	}

	l := left.(*obj.Integer).Val()
	r := right.(*obj.Integer).Val()

	if r == 0 {
		return obj.NewError("can't divide by 0")
	}
	return env.Set(name, obj.NewInteger(l%r))
}

func (m ModAssign) String() string {
	return fmt.Sprintf("(%v %%= %v)", m.l, m.r)
}

func (m ModAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: m.l, r: Mod{l: m.l, r: m.r, pos: m.pos}, pos: m.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (m ModAssign) IsConstExpression() bool {
	return false
}
