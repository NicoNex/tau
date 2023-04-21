package ast

import (
	"errors"
	"fmt"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type Call struct {
	Fn   Node
	Args []Node
	pos  int
}

func NewCall(fn Node, args []Node, pos int) Node {
	return Call{
		Fn:   fn,
		Args: args,
		pos:  pos,
	}
}

func (c Call) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.Call: not a constant expression")
}

func (c Call) String() string {
	var args []string

	for _, a := range c.Args {
		args = append(args, a.String())
	}
	return fmt.Sprintf("%v(%s)", c.Fn, strings.Join(args, ", "))
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
	if position, err = c.Fn.Compile(comp); err != nil {
		return
	}

	for _, a := range c.Args {
		if position, err = a.Compile(comp); err != nil {
			return
		}
	}

	position = comp.Emit(code.OpCall, len(c.Args))
	comp.Bookmark(c.pos)
	return
}

func (c Call) IsConstExpression() bool {
	return false
}
