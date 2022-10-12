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

func NewDivideAssign(l, r Node) Node {
	return DivideAssign{
		l:   l,
		r:   r,
		pos: 0,
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

	if !assertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '/=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '/=' for type %v", right.Type())
	}

	if gs, ok := left.(obj.GetSetter); ok {
		leftFl, rightFl := toFloat(gs.Object(), right)
		l := leftFl.(*obj.Float).Val()
		r := rightFl.(*obj.Float).Val()
		return gs.Set(obj.NewFloat(l / r))
	}

	leftFl, rightFl := toFloat(left, right)
	l := leftFl.(*obj.Float).Val()
	r := rightFl.(*obj.Float).Val()
	return env.Set(name, obj.NewFloat(l/r))
}

func (d DivideAssign) String() string {
	return fmt.Sprintf("(%v /= %v)", d.l, d.r)
}

func (d DivideAssign) Compile(c *compiler.Compiler) (position int, err error) {
	n := Assign{l: d.l, r: Divide{l: d.l, r: d.r, pos: d.pos}, pos: d.pos}
	return n.Compile(c)
}

func (d DivideAssign) IsConstExpression() bool {
	return false
}
