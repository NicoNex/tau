package ast

import (
	"fmt"

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
		lft = i.left.Eval(env)
		idx = i.index.Eval(env)
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

		if int(i) >= len(l) {
			return obj.NewError("intex out of range")
		}
		return l[i]

	case assertTypes(lft, obj.StringType) && assertTypes(idx, obj.IntType):
		s := lft.(*obj.String)
		i := idx.(*obj.Integer).Val()

		if int(i) >= len(*s) {
			return obj.NewError("intex out of range")
		}
		return obj.NewString(string(string(*s)[i]))

	case assertTypes(lft, obj.MapType) && assertTypes(idx, obj.IntType, obj.FloatType, obj.StringType, obj.BoolType):
		m := lft.(obj.Map)
		k := idx.(obj.Hashable)
		return m.Get(k.KeyHash()).Value

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
