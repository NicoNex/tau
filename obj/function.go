package obj

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/code"
)

type Function struct {
	Params       []string
	Body         interface{}
	Env          *Env
	Instructions code.Instructions
}

func NewFunction(params []string, env *Env, body interface{}) Object {
	return &Function{Params: params, Body: body, Env: env}
}

func NewFunctionCompiled(i code.Instructions) Object {
	return &Function{Instructions: i}
}

func (f Function) Type() Type {
	return FunctionType
}

func (f Function) String() string {
	return fmt.Sprintf("fn(%s) { %v }", strings.Join(f.Params, ", "), f.Body)
}
