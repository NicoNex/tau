package obj

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/internal/code"
)

type Evaluable interface {
	Eval(*Env) Object
}

type Function struct {
	Params []string
	Body   Evaluable
	Env    *Env
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
}

func NewFunctionCompiled(i code.Instructions, nLocals, nParams int) Object {
	return &CompiledFunction{
		Instructions: i,
		NumLocals:    nLocals,
		NumParams:    nParams,
	}
}

func (f CompiledFunction) Type() Type {
	return FunctionType
}

func (f CompiledFunction) String() string {
	return "<compiled function>"
}
