package tauimport

import (
	"os"

	"github.com/NicoNex/tau/obj"
	"github.com/NicoNex/tau/parser"
)

func resolve(name string) string {
	return name
}

func EvalImport(name string) obj.Object {
	b, err := os.ReadFile(resolve(name))
	if err != nil {
		return obj.NewError("import error: %w", err.Error())
	}

	tree, errs := parser.Parse(string(b))
	if len(errs) > 0 {
		return obj.NewError(
			"import: multiple errors in module %q:\n  %s",
			name,
			strings.Join(errs, "\n  "),
		)
	}

	env := obj.NewEnv()
	tree.Eval(env)

	return &obj.Class{env}
}
