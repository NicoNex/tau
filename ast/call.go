package ast

import (
	"fmt"
	"strings"
	"tau/obj"
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

	if fn, ok := fnObj.(*obj.Function); ok {
		var args []obj.Object

		for _, a := range c.args {
			args = append(args, a.Eval(env))
		}

		extEnv := extendEnv(fn, args)
		return unwrapReturn(fn.Body.(Node).Eval(extEnv))
	}
	return obj.NewError("%q object is not callable", fnObj.Type())
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
	if ret, ok := o.(*obj.Return); ok {
		return ret.Val()
	}
	return o
}
