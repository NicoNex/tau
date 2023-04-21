package ast

import (
	"errors"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type Continue struct{}

func NewContinue() Continue {
	return Continue{}
}

func (c Continue) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.ConcurrentCall: not a constant expression")
}

func (c Continue) String() string {
	return "break"
}

func (c Continue) Compile(comp *compiler.Compiler) (position int, err error) {
	return comp.Emit(code.OpJump, compiler.ContinuePlaceholder), nil
}

func (c Continue) IsConstExpression() bool {
	return false
}
