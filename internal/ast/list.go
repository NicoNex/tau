package ast

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type List []Node

func NewList(elements ...Node) Node {
	var ret List

	for _, e := range elements {
		ret = append(ret, e)
	}
	return ret
}

func (l List) Eval(env *obj.Env) obj.Object {
	var elements []obj.Object

	for _, e := range l {
		v := obj.Unwrap(e.Eval(env))
		if takesPrecedence(v) {
			return v
		}
		elements = append(elements, v)
	}
	return obj.NewList(elements...)
}

func (l List) String() string {
	var elements []string

	for _, e := range l {
		if s, ok := e.(String); ok {
			elements = append(elements, s.Quoted())
		} else {
			elements = append(elements, e.String())
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (l List) Compile(c *compiler.Compiler) (position int, err error) {
	for _, n := range l {
		if position, err = n.Compile(c); err != nil {
			return
		}
	}
	position = c.Emit(code.OpList, len(l))
	return
}
