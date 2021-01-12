package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type DivideAssign struct {
	l Node
	r Node
}

func NewDivideAssign(l, r Node) Node {
	return DivideAssign{l, r}
}

func (d DivideAssign) Eval(env *obj.Env) obj.Object {
	return obj.NullObj
}

func (d DivideAssign) String() string {
	return fmt.Sprintf("(%v /= %v)", d.l, d.r)
}
