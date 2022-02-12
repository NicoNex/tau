package ast

import (
	"strings"

	"github.com/NicoNex/tau/obj"
)

type Block []Node

func NewBlock() Block {
	return Block([]Node{})
}

func (b Block) Eval(env *obj.Env) obj.Object {
	var res obj.Object

	for _, n := range b {
		res = obj.Unwrap(n.Eval(env))
		if res != nil && takesPrecedence(res) {
			return res
		}
	}
	return res
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
