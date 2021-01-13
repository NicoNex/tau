package ast

import (
	"fmt"
	"github.com/NicoNex/tau/obj"
)

type PlusAssign struct {
	l Node
	r Node
}

func NewPlusAssign(l, r Node) Node {
	return PlusAssign{l, r}
}

func (p PlusAssign) Eval(env *obj.Env) obj.Object {
	var name string
	var left = p.l.Eval(env)
	var right = p.r.Eval(env)

	if ident, ok := p.l.(Identifier); ok {
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

	if !assertTypes(left, obj.INT, obj.FLOAT, obj.STRING) {
		return obj.NewError("unsupported operator '+=' for type %v", left.Type())
	}
	if !assertTypes(right, obj.INT, obj.FLOAT, obj.STRING) {
		return obj.NewError("unsupported operator '+=' for type %v", right.Type())
	}

	switch {
		case assertTypes(left, obj.STRING) && assertTypes(right, obj.STRING):
			l := left.(*obj.String).Val()
			r := right.(*obj.String).Val()
			env.Set(name, obj.NewString(l + r))

		case assertTypes(left, obj.INT) && assertTypes(right, obj.INT):
			l := left.(*obj.Integer).Val()
			r := right.(*obj.Integer).Val()
			env.Set(name, obj.NewInteger(l + r))

		case assertTypes(left, obj.FLOAT, obj.INT) && assertTypes(right, obj.FLOAT, obj.INT):
			left, right = toFloat(left, right)
			l := left.(*obj.Float).Val()
			r := right.(*obj.Float).Val()
			env.Set(name, obj.NewFloat(l + r))

		default:
			return obj.NewError(
				"invalid operation %v += %v (wrong types %v and %v)",
				left, right, left.Type(), right.Type(),
			)
	}

	return obj.NullObj
}

func (p PlusAssign) String() string {
	return fmt.Sprintf("(%v += %v)", p.l, p.r)
}
