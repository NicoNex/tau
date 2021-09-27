package ast

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/obj"
)

type In struct {
	l Node
	r Node
}

func NewIn(l, r Node) Node {
	return In{l, r}
}

func (i In) Eval(env *obj.Env) obj.Object {
	var (
		left  = unwrap(i.l.Eval(env))
		right = unwrap(i.r.Eval(env))
	)

	if isError(left) {
		return left
	}
	if isError(right) {
		return right
	}

	if !assertTypes(left, obj.INT, obj.FLOAT, obj.STRING, obj.BOOL, obj.NULL) {
		return obj.NewError("unsupported operator 'in' for type %v", left.Type())
	}
	if !assertTypes(right, obj.LIST, obj.STRING) {
		return obj.NewError("unsupported operator 'in' for type %v", right.Type())
	}

	switch {
	case assertTypes(left, obj.STRING) && assertTypes(right, obj.STRING):
		l := left.(*obj.String).Val()
		r := right.(*obj.String).Val()
		return obj.ParseBool(strings.Contains(r, l))

	case assertTypes(right, obj.LIST):
		for _, o := range right.(obj.List).Val() {
			if !assertTypes(left, o.Type()) {
				continue
			}
			if assertTypes(left, obj.BOOL, obj.NULL) && left == o {
				return obj.True
			}

			switch l := left.(type) {
			case *obj.String:
				r := o.(*obj.String)
				if l.Val() == r.Val() {
					return obj.True
				}

			case *obj.Integer:
				r := o.(*obj.Integer)
				if l.Val() == r.Val() {
					return obj.True
				}

			case *obj.Float:
				r := o.(*obj.Float)
				if l.Val() == r.Val() {
					return obj.True
				}
			}
		}
		return obj.False

	default:
		return obj.NewError(
			"invalid operation %v in %v (wrong types %v and %v)",
			left, right, left.Type(), right.Type(),
		)
	}
}

func (i In) String() string {
	return fmt.Sprintf("(%v in %v)", i.l, i.r)
}
