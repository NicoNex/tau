package ast

import (
	"fmt"
	"os"
	"strings"

	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type Import struct {
	name  Node
	parse func(string) (Node, []string)
}

func NewImport(name Node, parse func(string) (Node, []string)) Node {
	return &Import{
		name:  name,
		parse: parse,
	}
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

	b, err := os.ReadFile(string(*n))
	if err != nil {
		return obj.NewError("import error: %w", err.Error())
	}

	tree, errs := i.parse(string(b))
	if len(errs) > 0 {
		return obj.NewError(
			"import: multiple errors in module %q:\n  %s",
			name,
			strings.Join(errs, "\n  "),
		)
	}

	modEnv := obj.NewEnv()
	tree.Eval(modEnv)
	return obj.Class{Env: modEnv}
}

func (i Import) Compile(c *compiler.Compiler) (position int, err error) {
	if position, err = i.name.Compile(c); err != nil {
		return
	}
	return c.Emit(code.OpLoadModule), nil
}

func (i Import) String() string {
	return fmt.Sprintf("import(%q)", i.name.String())
}
