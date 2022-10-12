package ast

import (
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type PrefixMinus struct {
	n   Node
	pos int
}

func NewPrefixMinus(n Node, pos int) Node {
	return PrefixMinus{
		n:   n,
		pos: pos,
	}
}

func (p PrefixMinus) Eval(env *obj.Env) obj.Object {
	var value = obj.Unwrap(p.n.Eval(env))

	if takesPrecedence(value) {
		return value
	}

	switch v := value.(type) {
	case *obj.Integer:
		return obj.NewInteger(-v.Val())

	case *obj.Float:
		return obj.NewFloat(-v.Val())

	default:
		return obj.NewError("unsupported prefix operator '-' for type %v", value.Type())

	}
}

func (p PrefixMinus) String() string {
	return fmt.Sprintf("-%v", p.n)
}

func (p PrefixMinus) Compile(c *compiler.Compiler) (position int, err error) {
	if p.IsConstExpression() {
		return c.Emit(code.OpConstant, c.AddConstant(p.Eval(nil))), nil
	}

	if position, err = p.n.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpMinus), nil
}

func (p PrefixMinus) IsConstExpression() bool {
	return p.n.IsConstExpression()
}
