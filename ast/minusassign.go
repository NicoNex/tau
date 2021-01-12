package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type MinusAssign struct {
	l Node
	r Node
}

func NewMinusAssign(l, r Node) Node {
	return MinusAssign{l, r}
}

func (m MinusAssign) Eval(env *obj.Env) obj.Object {
	return obj.NullObj
}

func (m MinusAssign) String() string {
	return fmt.Sprintf("(%v -= %v)", m.l, m.r)
}
