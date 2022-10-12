package ast

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Function struct {
	params []Identifier
	body   Node
	Name   string
	pos    int
}

func NewFunction(params []Identifier, body Node, pos int) Node {
	return Function{
		params: params,
		body:   body,
		pos:    pos,
	}
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

	if f.Name != "" {
		c.DefineFunctionName(f.Name)
	}

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

	freeSymbols := c.FreeSymbols
	nLocals := c.NumDefs
	ins := c.LeaveScope()

	for _, s := range freeSymbols {
		position = c.LoadSymbol(s)
	}

	fn := obj.NewFunctionCompiled(ins, nLocals, len(f.params))
	position = c.Emit(code.OpClosure, c.AddConstant(fn), len(freeSymbols))
	c.Bookmark(f.pos)
	return
}

func (f Function) IsConstExpression() bool {
	return false
}
