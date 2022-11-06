package ast

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type ConcurrentCall struct {
	fn   Node
	args []Node
}

func NewConcurrentCall(fn Node, args []Node) Node {
	return ConcurrentCall{fn, args}
}

func (c ConcurrentCall) Eval(env *obj.Env) obj.Object {
	var fnObj = obj.Unwrap(c.fn.Eval(env))

	if takesPrecedence(fnObj) {
		return fnObj
	}

	switch fn := fnObj.(type) {
	case *obj.Function:
		var args []obj.Object

		if len(c.args) != len(fn.Params) {
			return obj.NewError(
				"wrong number of arguments: expected %d, got %d",
				len(fn.Params),
				len(c.args),
			)
		}

		for _, a := range c.args {
			o := obj.Unwrap(a.Eval(env))
			if takesPrecedence(o) {
				return o
			}
			args = append(args, o)
		}

		go fn.Body.(Node).Eval(extendEnv(fn, args))
		return obj.NullObj

	case obj.Builtin:
		var args []obj.Object

		for _, a := range c.args {
			args = append(args, obj.Unwrap(a.Eval(env)))
		}
		go fn(args...)
		return obj.NullObj

	default:
		return obj.NewError("%q object is not callable", fnObj.Type())
	}
}

func (c ConcurrentCall) String() string {
	var args = make([]string, len(c.args))

	for i, a := range c.args {
		args[i] = a.String()
	}
	return fmt.Sprintf("tau %v(%s)", c.fn, strings.Join(args, ", "))
}

func (c ConcurrentCall) Compile(comp *compiler.Compiler) (position int, err error) {
	if position, err = c.fn.Compile(comp); err != nil {
		return
	}

	for _, a := range c.args {
		if position, err = a.Compile(comp); err != nil {
			return
		}
	}

	return comp.Emit(code.OpConcurrentCall, len(c.args)), nil
}

func (ConcurrentCall) IsConstExpression() bool {
	return false
}
