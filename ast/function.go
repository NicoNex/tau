package ast

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type Function struct {
	params []Identifier
	body   Node
}

func NewFunction(params []Identifier, body Node) Node {
	return Function{params, body}
}

func (f Function) Eval(env *obj.Env) obj.Object {
	var params []string

	for _, p := range f.params {
		params = append(params, p.String())
	}

	return obj.NewFunction(params, env, f.body)
}

func (f Function) String() string {
	var params []string

	for _, p := range f.params {
		params = append(params, p.String())
	}
	return fmt.Sprintf("fn(%s) { %v }", strings.Join(params, ", "), f.body)
}

func (f Function) Compile(c *compiler.Compiler) (position int, err error) {
	c.EnterScope()

	for _, p := range f.params {
		c.Define(p.String())
	}

	if position, err = f.body.Compile(c); err != nil {
		return
	}

	if c.LastIs(code.OpPop) {
		c.ReplaceLastPopWithReturn()
	}
	if !c.LastIs(code.OpReturnValue) {
		c.Emit(code.OpReturn)
	}

	nLocals := c.NumDefs
	ins := c.LeaveScope()
	fn := obj.NewFunctionCompiled(ins, nLocals, len(f.params))
	return c.Emit(code.OpConstant, c.AddConstant(fn)), nil
}
