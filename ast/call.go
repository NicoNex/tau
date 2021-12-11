package ast

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type Call struct {
	fn   Node
	args []Node
}

func NewCall(fn Node, args []Node) Node {
	return Call{fn, args}
}

func (c Call) Eval(env *obj.Env) obj.Object {
	var fnObj = c.fn.Eval(env)

	if takesPrecedence(fnObj) {
		return fnObj
	}

	switch fn := fnObj.(type) {
	case *obj.Function:
		var args []obj.Object

		if len(c.args) != len(fn.Params) {
			return obj.NewError(
				"wrong number of arguments, expected %d, got %d",
				len(fn.Params),
				len(c.args),
			)
		}

		for _, a := range c.args {
			o := a.Eval(env)
			if takesPrecedence(o) {
				return o
			}
			args = append(args, o)
		}

		extEnv := extendEnv(fn, args)
		return unwrapReturn(fn.Body.(Node).Eval(extEnv))

	case obj.Builtin:
		var args []obj.Object

		for _, a := range c.args {
			args = append(args, a.Eval(env))
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

func (c Call) Compile(comp *compiler.Compiler) int {
	return 0
}

func extendEnv(fn *obj.Function, args []obj.Object) *obj.Env {
	var env = obj.NewEnvWrap(fn.Env)

	for i, p := range fn.Params {
		env.Set(p, args[i])
	}
	return env
}

func unwrapReturn(o obj.Object) obj.Object {
	if ret, ok := o.(*obj.Return); ok {
		return ret.Val()
	}
	return o
}
