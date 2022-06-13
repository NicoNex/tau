package ast

import (
	"fmt"

	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type Index struct {
	left  Node
	index Node
}

func NewIndex(l, i Node) Node {
	return Index{l, i}
}

func (i Index) Eval(env *obj.Env) obj.Object {
	var (
		lft = obj.Unwrap(i.left.Eval(env))
		idx = obj.Unwrap(i.index.Eval(env))
	)

	if takesPrecedence(lft) {
		return lft
	}
	if takesPrecedence(idx) {
		return idx
	}

	switch {
	case assertTypes(lft, obj.ListType) && assertTypes(idx, obj.IntType):
		l := lft.(obj.List)
		i := idx.(*obj.Integer).Val()

		return &obj.GetSetterImpl{
			GetFunc: func() (obj.Object, bool) {
				if i < 0 || int(i) >= len(l) {
					return obj.NewError("intex out of range"), false
				}
				return l[i], true
			},

			SetFunc: func(o obj.Object) obj.Object {
				if i < 0 || int(i) >= len(l) {
					return obj.NewError("intex out of range")
				}

				l[i] = o
				return o
			},
		}

	case assertTypes(lft, obj.StringType) && assertTypes(idx, obj.IntType):
		s := lft.(*obj.String)
		i := idx.(*obj.Integer).Val()

		if i < 0 || int(i) >= len(*s) {
			return obj.NewError("intex out of range")
		}
		return obj.NewString(string(string(*s)[i]))

	case assertTypes(lft, obj.MapType) && assertTypes(idx, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType):
		m := lft.(obj.Map)
		k := idx.(obj.Hashable)
		return &obj.GetSetterImpl{
			GetFunc: func() (obj.Object, bool) {
				v := m.Get(k.KeyHash()).Value
				return v, v != obj.NullObj
			},

			SetFunc: func(o obj.Object) obj.Object {
				m.Set(k.KeyHash(), obj.MapPair{Key: idx, Value: o})
				return o
			},
		}

	default:
		return obj.NewError(
			"invalid index operator for types %v and %v",
			lft.Type(),
			idx.Type(),
		)
	}
}

func (i Index) String() string {
	return fmt.Sprintf("%v[%v]", i.left, i.index)
}

func (i Index) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = i.left.Compile(c); err != nil {
		return
	}
	if position, err = i.index.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpIndex), nil
}
