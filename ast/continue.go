package ast

import "github.com/NicoNex/tau/obj"

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
