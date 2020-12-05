package ast

import (
	"fmt"
	"tau/obj"
)

type IfElse struct {
	cond Node
	body Node
	altern Node
}

func (i IfElse) Eval() obj.Object {
	var cond = i.cond.Eval()

	switch c := cond.(type) {
	case *obj.Boolean:
		if c.Val() {
			return body.Eval()
		}
		return altern.Eval()

	case *obj.Null:
		return obj.False

	default:
		return obj.True
	}
}

func (i IfElse) String() string {
	if i.altern != nil {
		return fmt.Sprintf("if %v { %v } else { %v }", i.cond, i.body, i.altern)
	}
	return fmt.Sprintf("if %v { %v }", i.cond, i.body)
}
