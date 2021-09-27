package ast

import (
	"fmt"

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
		name        string
		isContainer bool
		left        = d.l.Eval(env)
		right       = unwrap(d.r.Eval(env))
	)

	if ident, ok := d.l.(Identifier); ok {
		name = ident.String()
	} else if _, isContainer = left.(*obj.Container); !isContainer {
		return obj.NewError("cannot assign to literal")
	}

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '/=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.IntType, obj.FloatType) {
		return obj.NewError("unsupported operator '/=' for type %v", right.Type())
	}

	leftFl, rightFl := toFloat(unwrap(left), right)
	l := leftFl.(*obj.Float).Val()
	r := rightFl.(*obj.Float).Val()

	if isContainer {
		return left.(*obj.Container).Set(obj.NewFloat(l / r))
	}
	return env.Set(name, obj.NewFloat(l/r))
}

func (d DivideAssign) String() string {
	return fmt.Sprintf("(%v /= %v)", d.l, d.r)
}
