package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type Index struct {
	list  Node
	index Node
}

func NewIndex(l, i Node) Node {
	return Index{l, i}
}

func (i Index) Eval(env *obj.Env) obj.Object {
	var lst = i.list.Eval(env)
	var idx = i.index.Eval(env)

	if isError(lst) {
		return lst
	}
	if isError(idx) {
		return idx
	}

	switch {
	case assertTypes(lst, obj.LIST) && assertTypes(idx, obj.INT):
		l := lst.(obj.List)
		i := idx.(*obj.Integer).Val()

		if int(i) >= len(l) {
			return obj.NewError("intex out of range")
		}
		return l[i]

	case assertTypes(lst, obj.STRING) && assertTypes(idx, obj.INT):
		s := lst.(*obj.String)
		i := idx.(*obj.Integer).Val()

		if int(i) >= len(*s) {
			return obj.NewError("intex out of range")
		}
		return obj.NewString(string(string(*s)[i]))

	default:
		return obj.NewError(
			"invalid index operator for types %v and %v",
			lst.Type(),
			idx.Type(),
		)
	}
}

func (i Index) String() string {
	return fmt.Sprintf("%v[%v]", i.list, i.index)
}
