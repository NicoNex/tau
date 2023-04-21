package ast

import (
	"errors"
	"fmt"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type List []Node

func NewList(elements ...Node) Node {
	var ret List

	for _, e := range elements {
		ret = append(ret, e)
	}
	return ret
}

// TODO: optimise this for the case where all the list elements are constant expressions.
func (l List) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.Index: not a constant expression")
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

func (l List) IsConstExpression() bool {
	return false
}
