package ast

import "tau/obj"

type String string

func NewString(s string) Node {
	return String(s)
}

func (s String) Eval() obj.Object {
	return obj.NewString(string(s))
}

func (s String) String() string {
	return string(s)
}
