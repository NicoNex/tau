package ast

import "github.com/NicoNex/tau/obj"

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
