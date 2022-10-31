package ast

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type Call struct {
	fn   Node
	args []Node
}

func NewCall(fn Node, args []Node) Node {
	return Call{fn, args}
}

func (c Call) Eval(env *obj.Env) obj.Object {
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

		extEnv := extendEnv(fn, args)
		return unwrapReturn(fn.Body.Eval(extEnv))

	case obj.Builtin:
		var args []obj.Object

		for _, a := range c.args {
			args = append(args, obj.Unwrap(a.Eval(env)))
		}
		return fn(args...)

	default:
		return obj.NewError("%q object is not callable", fnObj.Type())
	}
}

func (c Call) String() string {
	var args []string

	for _, a := range c.args {
		args = append(args, a.String())
	}
	return fmt.Sprintf("%v(%s)", c.fn, strings.Join(args, ", "))
}

func extendEnv(fn *obj.Function, args []obj.Object) *obj.Env {
	var env = obj.NewEnvWrap(fn.Env)

	for i, p := range fn.Params {
		env.Set(p, args[i])
	}
	return env
}

func unwrapReturn(o obj.Object) obj.Object {
	if ret, ok := o.(obj.Return); ok {
		return ret.Val()
	}
	return o
}

func (c Call) Compile(comp *compiler.Compiler) (position int, err error) {
	if position, err = c.fn.Compile(comp); err != nil {
		return
	}

	for _, a := range c.args {
		if position, err = a.Compile(comp); err != nil {
			return
		}
	}

	return comp.Emit(code.OpCall, len(c.args)), nil
}

func (c Call) IsConstExpression() bool {
	return false
}
