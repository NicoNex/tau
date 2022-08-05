package ast

import (
	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Continue struct{}

func NewContinue() Continue {
	return Continue{}
}

func (c Continue) Eval(_ *obj.Env) obj.Object {
	return obj.ContinueObj
}

func (c Continue) String() string {
	return "break"
}

func (c Continue) Compile(comp *compiler.Compiler) (position int, err error) {
	return comp.Emit(code.OpJump, compiler.ContinuePlaceholder), nil
}
