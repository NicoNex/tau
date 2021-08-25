package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type For struct {
	before Node
	after  Node
	cond   Node
	body   Node
}

func NewFor(cond, body, before, after Node) Node {
	return For{before, after, cond, body}
}

func (f For) Eval(env *obj.Env) obj.Object {
	if f.before != nil {
		f.before.Eval(env)
	}
	for isTruthy(f.cond.Eval(env)) {
		if o := f.body.Eval(env); isError(o) {
			return o
		}
		if f.after != nil {
			f.after.Eval(env)
		}
	}
	return obj.NullObj
}

func (f For) String() string {
	return fmt.Sprintf("for %v { %v }", f.cond, f.body)
}
