package ast

import (
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type RawString string

func NewRawString(s string) Node {
	return RawString(s)
}

func (r RawString) Eval(env *obj.Env) obj.Object {
	return obj.String(string(r))
}

func (r RawString) String() string {
	return string(r)
}

func (r RawString) Quoted() string {
	var buf strings.Builder

	buf.WriteRune('`')
	buf.WriteString(string(r))
	buf.WriteRune('`')
	return buf.String()
}

func (r RawString) Compile(c *compiler.Compiler) (position int, err error) {
	return c.Emit(code.OpConstant, c.AddConstant(c.NewString(string(r)))), nil
}

func (r RawString) IsConstExpression() bool {
	return true
}
