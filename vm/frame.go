package vm

import (
	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/obj"
)

type Frame struct {
	fn *obj.Function
	ip int
}

func NewFrame(fn *obj.Function) *Frame {
	return &Frame{fn: fn, ip: -1}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
