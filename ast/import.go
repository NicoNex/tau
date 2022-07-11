package ast

import (
	"fmt"

	// "github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
	"github.com/NicoNex/tau/tauimport"
)

type Import struct {
	name Node
}

func NewImport(name Node) Node {
	return &Import{name: name}
}

func (i Import) Eval(env *obj.Env) obj.Object {
	var name = obj.Unwrap(i.name.Eval(env))

	if takesPrecedence(name) {
		return name
	}

	n, ok := name.(*obj.String)
	if !ok {
		return obj.NewError("import: expected string but got %v", name.Type())
	}

	return tauimport.EvalImport(*n)
}

func (i Import) Compile(comp *compiler.Compiler) (position int, err error) {
	return 0, nil
}

func (i Import) String() string {
	return fmt.Sprintf("import(%q)", i.name.String())
}
