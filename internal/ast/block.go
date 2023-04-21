package ast

import (
	"errors"
	"strings"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/vm/cvm/cobj"
)

type Block []Node

func NewBlock() Block {
	return Block([]Node{})
}

func (b Block) Eval() (cobj.Object, error) {
	return cobj.NullObj, errors.New("ast.Block: not a constant expression")
}

func (b Block) String() string {
	var nodes []string

	for _, n := range b {
		nodes = append(nodes, n.String())
	}
	return strings.Join(nodes, "; ")
}

func (b *Block) Add(n Node) {
	*b = append(*b, n)
}

func (b Block) Compile(c *compiler.Compiler) (position int, err error) {
	for _, n := range b {
		if position, err = n.Compile(c); err != nil {
			return
		}

		if _, isReturn := n.(Return); !isReturn {
			position = c.Emit(code.OpPop)
		}
	}
	return
}

func (b Block) IsConstExpression() bool {
	return false
}
