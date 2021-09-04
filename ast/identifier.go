package ast

import "github.com/NicoNex/tau/obj"

type Identifier string

func NewIdentifier(name string) Node {
	return Identifier(name)
}

func (i Identifier) Eval(env *obj.Env) obj.Object {
	if c, ok := env.Get(string(i)); ok {
		return c.Object()
	} else if o, ok := obj.Builtins[string(i)]; ok {
		return o
	}

	return obj.NewError("name %q is not defined", i)
}

func (i Identifier) String() string {
	return string(i)
}
