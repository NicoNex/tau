package vm

import (
	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/obj"
)

type Frame struct {
	cl          *obj.Closure
	ip          int
	basePointer int
}

func NewFrame(cl *obj.Closure, basePointer int) *Frame {
	return &Frame{
		cl:          cl,
		ip:          -1,
		basePointer: basePointer,
	}
}

func (f *Frame) Instructions() code.Instructions {
	return f.cl.Fn.Instructions
}
