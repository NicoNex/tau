package ast

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type In struct {
	l   Node
	r   Node
	pos int
}

func NewIn(l, r Node, pos int) Node {
	return In{
		l:   l,
		r:   r,
		pos: pos,
	}
}

func (i In) Eval(env *obj.Env) obj.Object {
	var (
		left  = obj.Unwrap(i.l.Eval(env))
		right = obj.Unwrap(i.r.Eval(env))
	)

	if takesPrecedence(left) {
		return left
	}
	if takesPrecedence(right) {
		return right
	}

	if !obj.AssertTypes(left, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType, obj.NullType) {
		return obj.NewError("unsupported operator 'in' for type %v", left.Type())
	}
	if !obj.AssertTypes(right, obj.ListType, obj.StringType) {
		return obj.NewError("unsupported operator 'in' for type %v", right.Type())
	}

	switch {
	case obj.AssertTypes(left, obj.StringType) && obj.AssertTypes(right, obj.StringType):
		l := left.(obj.String).Val()
		r := right.(obj.String).Val()
		return obj.ParseBool(strings.Contains(r, l))

	case obj.AssertTypes(right, obj.ListType):
		for _, o := range right.(obj.List).Val() {
			if !obj.AssertTypes(left, o.Type()) {
				continue
			}
			if obj.AssertTypes(left, obj.BoolType, obj.NullType) && left == o {
				return obj.True
			}

			switch l := left.(type) {
			case obj.String:
				r := o.(obj.String)
				if l.Val() == r.Val() {
					return obj.True
				}

			case obj.Integer:
				r := o.(obj.Integer)
				if l.Val() == r.Val() {
					return obj.True
				}

			case obj.Float:
				r := o.(obj.Float)
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

func (i In) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = i.l.Compile(c); err != nil {
		return
	}
	if position, err = i.r.Compile(c); err != nil {
		return
	}
	position = c.Emit(code.OpIn)
	c.Bookmark(i.pos)
	return
}

func (i In) IsConstExpression() bool {
	return false
}
