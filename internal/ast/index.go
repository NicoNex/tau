package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Index struct {
	left  Node
	index Node
	pos   int
}

func NewIndex(l, i Node, pos int) Node {
	return Index{
		left:  l,
		index: i,
		pos:   pos,
	}
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
	case obj.AssertTypes(lft, obj.ListType) && obj.AssertTypes(idx, obj.IntType):
		l := lft.(obj.List)
		i := idx.(obj.Integer)

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

	case obj.AssertTypes(lft, obj.StringType) && obj.AssertTypes(idx, obj.IntType):
		s := lft.(obj.String)
		i := idx.(obj.Integer)

		if i < 0 || int(i) >= len(s) {
			return obj.NewError("intex out of range")
		}
		return obj.String(string(string(s)[i]))

	case obj.AssertTypes(lft, obj.BytesType) && obj.AssertTypes(idx, obj.IntType):
		b := lft.(obj.Bytes)
		i := idx.(obj.Integer)

		if i < 0 || int(i) >= len(b) {
			return obj.NewError("intex out of range")
		}
		return obj.Integer(int64(b[i]))

	case obj.AssertTypes(lft, obj.MapType) && obj.AssertTypes(idx, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType):
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
	position = c.Emit(code.OpIndex)
	c.Bookmark(i.pos)
	return
}

func (i Index) IsConstExpression() bool {
	return false
}
