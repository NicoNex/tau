package obj

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/code"
)

type Function struct {
	Params       []string
	Body         any
	Env          *Env
	Instructions code.Instructions
	NumLocals    int
	NumParams    int
}

func NewFunction(params []string, env *Env, body any) Object {
	return &Function{Params: params, Body: body, Env: env}
}

func NewFunctionCompiled(i code.Instructions, nLocals, nParams int) Object {
	return &Function{
		Instructions: i,
		NumLocals:    nLocals,
		NumParams:    nParams,
	}
}

func (f Function) Type() Type {
	return FunctionType
}

func (f Function) String() string {
	return fmt.Sprintf("fn(%s) { %v }", strings.Join(f.Params, ", "), f.Body)
}
