package ast

import (
	"strconv"
	"tau/obj"
)

type Integer int64

// Maybe remove this.
func NewInteger(i int64) Node {
	return Integer(i)
}

func (i Integer) Eval() obj.Object {
	return obj.NewInteger(int64(i))
}

func (i Integer) String() string {
	return strconv.ParseInt(int64(i), 64)
}
