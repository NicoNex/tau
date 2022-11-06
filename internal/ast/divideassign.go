package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type DivideAssign struct {
	l   Node
	r   Node
	pos int
}

func NewDivideAssign(l, r Node, pos int) Node {
	return DivideAssign{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (d DivideAssign) Eval(env *obj.Env) obj.Object {
	var (
		name  string
		left  = d.l.Eval(env)
		right = obj.Unwrap(d.r.Eval(env))
	)

	if ident, ok := d.l.(Identifier); ok {
		name = ident.String()
	}

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '/=' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '/=' for type %v", right.Type())
	}

	if gs, ok := left.(obj.GetSetter); ok {
		leftFl, rightFl := obj.ToFloat(gs.Object(), right)
		l := leftFl.(obj.Float)
		r := rightFl.(obj.Float)
		return gs.Set(obj.Float(l / r))
	}

	leftFl, rightFl := obj.ToFloat(left, right)
	l := leftFl.(obj.Float)
	r := rightFl.(obj.Float)
	return env.Set(name, obj.Float(l/r))
}

func (d DivideAssign) String() string {
	return fmt.Sprintf("(%v /= %v)", d.l, d.r)
}

func (d DivideAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: d.l, r: Divide{l: d.l, r: d.r, pos: d.pos}, pos: d.pos}
	position, err = n.Compile(c)
	c.Bookmark(n.pos)
	return
}

func (d DivideAssign) IsConstExpression() bool {
	return false
}
