package ast

import (
	"fmt"
	"tau/obj"
)

type Return struct {
	v Node
}

func NewReturn(n Node) Node {
	return Return{n}
}

func (r Return) Eval() obj.Object {
	return obj.NewReturn(r.v.Eval())
}

func (r Return) String() string {
	return fmt.Sprintf("return %v;", r.v)
}
