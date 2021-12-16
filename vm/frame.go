package vm

import (
	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/obj"
)

type Frame struct {
	fn          *obj.Function
	ip          int
	basePointer int
}

func NewFrame(fn *obj.Function, basePointer int) *Frame {
	return &Frame{
		fn:          fn,
		ip:          -1,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.fn.Instructions
}
