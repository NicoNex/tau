package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type For struct {
	cond Node
	body Node
}

func NewFor(cond, body Node) Node {
	return For{cond, body}
}

func (f For) shouldContinue(env *obj.Env) bool {
	return false
}

func (f For) Eval(env *obj.Env) obj.Object {
	return obj.NullObj
}

func (f For) String() string {
	return fmt.Sprintf("for %v { %v }", f.cond, f.body)
}
