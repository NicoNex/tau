package ast

import (
	"fmt"
<<<<<<< HEAD

=======
>>>>>>> f638442 (Implementing plus assign (+=) operator)
	"github.com/NicoNex/tau/obj"
)

type PlusAssign struct {
	l Node
	r Node
}

func NewPlusAssign(l, r Node) Node {
	return PlusAssign{l, r}
}

<<<<<<< HEAD
func (p PlusAssign) Eval(env *obj.Env) obj.Object {
=======
// TODO: fix the bug in case a builtin function returns an error.
func (p PlusAssign) Eval(env *obj.Env) obj.Object {
	var left = p.l.Eval(env)
	var right = p.r.Eval(env)

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
			env.Set(left.String(), obj.NewString(l + r))
			

		case assertTypes(left, obj.INT) && assertTypes(right, obj.INT):
			l := left.(*obj.Integer).Val()
			r := right.(*obj.Integer).Val()
			env.Set(left.String(), obj.NewInteger(l + r))

		case assertTypes(left, obj.FLOAT, obj.INT) && assertTypes(right, obj.FLOAT, obj.INT):
			left, right = toFloat(left, right)
			l := left.(*obj.Float).Val()
			r := right.(*obj.Float).Val()
			env.Set(left.String(), obj.NewFloat(l + r))

		default:
			return obj.NewError(
				"invalid operation %v += %v (wrong types %v and %v)",
				left, right, left.Type(), right.Type(),
			)
	}

>>>>>>> f638442 (Implementing plus assign (+=) operator)
	return obj.NullObj
}

func (p PlusAssign) String() string {
	return fmt.Sprintf("(%v += %v)", p.l, p.r)
}
