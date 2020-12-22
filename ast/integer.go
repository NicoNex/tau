package ast

import (
	"strconv"

	"github.com/NicoNex/tau/obj"
)

type Integer int64

func NewInteger(i int64) Node {
	return Integer(i)
}

func (i Integer) Eval(env *obj.Env) obj.Object {
	return obj.NewInteger(int64(i))
}

func (i Integer) String() string {
	return strconv.FormatInt(int64(i), 10)
}
