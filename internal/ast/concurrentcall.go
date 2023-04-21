package ast

import (
	"errors"
	"fmt"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type ConcurrentCall struct {
	fn   Node
	args []Node
}

func NewConcurrentCall(fn Node, args []Node) Node {
	return ConcurrentCall{fn, args}
}

func (c ConcurrentCall) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.ConcurrentCall: not a constant expression")
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
