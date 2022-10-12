package ast

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Import struct {
	name  Node
	parse func(string) (Node, []string)
	pos   int
}

func NewImport(name Node, parse func(string) (Node, []string)) Node {
	return &Import{
		name:  name,
		parse: parse,
		pos:   0,
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

	path, err := obj.ImportLookup(filepath.Join(env.Dir(), string(*n)))
	if err != nil {
		return obj.NewError("import: %v", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return obj.NewError("import: %v", err)
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
	dir, _ := filepath.Split(path)
	modEnv.SetDir(dir)
	tree.Eval(modEnv)
	return modEnv.Module()
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

func (i Import) IsConstExpression() bool {
	return false
}
