package ast

import (
	"fmt"

	"github.com/NicoNex/tau/obj"
)

type IfExpr struct {
	cond   Node
	body   Node
	altern Node
}

func NewIfExpr(cond, body, alt Node) Node {
	return IfExpr{cond, body, alt}
}

func (i IfExpr) Eval(env *obj.Env) obj.Object {
	var cond = i.cond.Eval(env)

	if takesPrecedence(cond) {
		return cond
	}

	switch c := cond.(type) {
	case *obj.Boolean:
		if c.Val() {
			return i.body.Eval(env)
		}
		return i.alternative(env)

	case *obj.Null:
		return i.alternative(env)

	default:
		return i.body.Eval(env)
	}
}

func (i IfExpr) String() string {
	if i.altern != nil {
		return fmt.Sprintf("if %v { %v } else { %v }", i.cond, i.body, i.altern)
	}
	return fmt.Sprintf("if %v { %v }", i.cond, i.body)
}

func (i IfExpr) alternative(env *obj.Env) obj.Object {
	if i.altern != nil {
		return i.altern.Eval(env)
	}
	return obj.NullObj
}
