package ast

import "github.com/NicoNex/tau/obj"

type String string

func NewString(s string) Node {
	return String(s)
}

func (s String) Eval(env *obj.Env) obj.Object {
	return obj.NewString(string(s))
}

func (s String) String() string {
	return string(s)
}
