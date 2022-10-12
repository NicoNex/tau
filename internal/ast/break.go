package ast

import (
	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Break struct{}

func NewBreak() Break {
	return Break{}
}

func (b Break) Eval(_ *obj.Env) obj.Object {
	return obj.BreakObj
}

func (b Break) String() string {
	return "break"
}

func (b Break) Compile(c *compiler.Compiler) (position int, err error) {
	return c.Emit(code.OpJump, compiler.BreakPlaceholder), nil
}

func (b Break) IsConstExpression() bool {
	return false
}
