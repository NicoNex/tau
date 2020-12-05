package obj

import (
	"fmt"
	"strings"
)

type Function struct {
	Params []string
	Body   interface{}
	Env    *Env
}

func NewFunction(params []string, env *Env, body interface{}) Object {
	return &Function{params, body, env}
}

func (f Function) Type() Type {
	return FUNCTION
}

func (f Function) String() string {
	return fmt.Sprintf("fn(%s) { %v }", strings.Join(f.Params, ", "), f.Body)
}
