package ast

import (
	"strconv"

	"github.com/NicoNex/tau/obj"
)

type Float float64

func NewFloat(f float64) Node {
	return Float(f)
}

func (f Float) Eval(env *obj.Env) obj.Object {
	return obj.NewFloat(float64(f))
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}
