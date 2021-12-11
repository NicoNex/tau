package ast

import (
	"fmt"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type DivideAssign struct {
	l Node
	r Node
}

func NewDivideAssign(l, r Node) Node {
	return DivideAssign{l, r}
}

func (d DivideAssign) Eval(env *obj.Env) obj.Object {
	var (
		name  string
		left  = d.l.Eval(env)
		right = d.r.Eval(env)
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

	leftFl, rightFl := toFloat(left, right)
	l := leftFl.(*obj.Float).Val()
	r := rightFl.(*obj.Float).Val()
	return env.Set(name, obj.NewFloat(l/r))
}

func (d DivideAssign) String() string {
	return fmt.Sprintf("(%v /= %v)", d.l, d.r)
}

func (d DivideAssign) Compile(c *compiler.Compiler) (position int) {
	return 0
}
