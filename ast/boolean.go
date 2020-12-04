package ast

import "tau/obj"

type Boolean bool

func NewBoolean(b bool) Node {
	return Boolean(b)
}

func (b Boolean) Eval() obj.Object {
	return obj.ParseBool(bool(b))
}

func (b Boolean) String() string {
	if b {
		return "true"
	}
	return "false"
}
