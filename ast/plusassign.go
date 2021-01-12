package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type PlusAssign struct {
	l Node
	r Node
}

func NewPlusAssign(l, r Node) Node {
	return PlusAssign{l, r}
}

func (p PlusAssign) Eval(env *obj.Env) obj.Object {
	return obj.NullObj
}

func (p PlusAssign) String() string {
	return fmt.Sprintf("(%v += %v)", p.l, p.r)
}
