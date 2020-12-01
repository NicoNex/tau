package ast

import (
	"strconv"
	"tau/obj"
)

type Float float64

func NewFloat(f float64) Node {
	return Float(f)
}

func (f Float) Eval() obj.Object {
	return obj.NewFloat(float64(f))
}

func (f Float) String() string {
	return strconv.ParseFloat(float64(f), 'f', -1, 64)
}
