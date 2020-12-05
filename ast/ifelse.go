package ast

import (
	"fmt"
	"tau/obj"
)

type IfExpr struct {
	cond Node
	body Node
	altern Node
}

func NewIfExpr(cond, body, alt Node) Node {
	return IfExpr{cond, body, alt}
}

func (i IfExpr) Eval() obj.Object {
	var cond = i.cond.Eval()

	switch c := cond.(type) {
	case *obj.Boolean:
		if c.Val() {
			return i.body.Eval()
		}
		return i.alternative()

	case *obj.Null:
		return i.alternative()

	default:
		return i.body.Eval()
	}
}

func (i IfExpr) String() string {
	if i.altern != nil {
		return fmt.Sprintf("if %v { %v } else { %v }", i.cond, i.body, i.altern)
	}
	return fmt.Sprintf("if %v { %v }", i.cond, i.body)
}

func (i IfExpr) alternative() obj.Object {
	if i.altern != nil {
		return i.altern.Eval()
	}
	return obj.NullObj
}
