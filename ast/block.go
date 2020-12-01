package ast

import "strings"

type Block struct {
	nodes []Node
}

func (b Block) Eval() obj.Object {
	var res obj.Object

	for _, n := range nodes {
		res = n.Eval()

		if res != nil {
			typ := res.Type()
			if typ == obj.RETURN || typ == obj.ERROR {
				return res
			}
		}
	}
	return res
}

func (b Block) String() string {
	var nodes []string

	for _, n := range b.nodes {
		nodes = append(nodes, n.String())
	}
	return strings.Join(nodes, "; ")
}

func (b *Block) Add(n Node) {
	b.nodes = append(b.nodes, n)
}
