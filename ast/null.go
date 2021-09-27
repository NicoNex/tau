package ast

import "github.com/NicoNex/tau/obj"

type Null struct{}

func NewNull() Null {
	return Null{}
}

func (n Null) Eval(_ *obj.Env) obj.Object {
	return obj.NullObj
}

func (n Null) String() string {
	return "null"
}
