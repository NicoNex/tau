package obj

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/tauerr"
)

type Evaluable interface {
	Eval(*Env) Object
}

type Function struct {
	Body   Evaluable
	Env    *Env
	Params []string
}

func NewFunction(params []string, env *Env, body Evaluable) Object {
	return &Function{Params: params, Body: body, Env: env}
}

func (f Function) Type() Type {
	return FunctionType
}

func (f Function) String() string {
	return fmt.Sprintf("fn(%s) { %v }", strings.Join(f.Params, ", "), f.Body)
}

type CompiledFunction struct {
	Instructions code.Instructions
	NumLocals    int
	NumParams    int
	Bookmarks    []tauerr.Bookmark
}

func NewFunctionCompiled(i code.Instructions, nLocals, nParams int, bookmarks []tauerr.Bookmark) Object {
	return &CompiledFunction{
		Instructions: i,
		NumLocals:    nLocals,
		NumParams:    nParams,
		Bookmarks:    bookmarks,
	}
}

func (c CompiledFunction) Type() Type {
	return FunctionType
}

func (c CompiledFunction) String() string {
	return "<compiled function>"
}
