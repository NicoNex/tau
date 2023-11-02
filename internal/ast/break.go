package ast

import (
	"errors"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Break struct{}

func NewBreak() Break {
	return Break{}
}

func (b Break) Eval() (obj.Object, error) {
	return obj.NullObj, errors.New("ast.Break: not a constant expression")
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
