package ast

import (
	"fmt"
	"strings"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
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
		v := e.Eval(env)
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
	return 0, nil
}
