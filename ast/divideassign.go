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
	var name string
	var left = d.l.Eval(env)
	var right = d.r.Eval(env)

	if ident, ok := d.l.(Identifier); ok {
    	name = ident.String()
	} else {
    	return obj.NewError("cannot assign to literal")
	}

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '/=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT, obj.FLOAT) {
		return obj.NewError("unsupported operator '/=' for type %v", right.Type())
	}

	left, right = toFloat(left, right)
	l := left.(*obj.Float).Val()
	r := right.(*obj.Float).Val()
	env.Set(name, obj.NewFloat(l / r))

	return obj.NullObj
}

func (d DivideAssign) String() string {
	return fmt.Sprintf("(%v /= %v)", d.l, d.r)
}
