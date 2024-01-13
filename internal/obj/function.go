package obj

import (
	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/tauerr"
)

type Function struct {
	Instructions code.Instructions
	NumLocals    int
	NumParams    int
	Bookmarks    []tauerr.Bookmark
}

func NewFunction(i code.Instructions, nLocals, nParams int, bookmarks []tauerr.Bookmark) Object {
	return Function{
		Instructions: i,
		NumLocals:    nLocals,
		NumParams:    nParams,
		Bookmarks:    bookmarks,
	}
}

func (c Function) Type() Type {
	return FunctionType
}

func (c Function) String() string {
	return "<function>"
}
