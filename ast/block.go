package ast

import (
	"strings"

	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
)

type Block []Node

func NewBlock() Block {
	return Block([]Node{})
}

func (b Block) Eval(env *obj.Env) obj.Object {
	var res obj.Object

	for _, n := range b {
		res = n.Eval(env)
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

func (b Block) Compile(c *compiler.Compiler) (position int, err error) {
	for _, n := range b {
		if position, err = n.Compile(c); err != nil {
			return
		}

		if canPop(n) {
			position = c.Emit(code.OpPop)
		}
	}
	return
}

func canPop(n Node) bool {
	switch n.(type) {
	case Assign,
		BitwiseAndAssign,
		BitwiseOrAssign,
		BitwiseShiftLeftAssign,
		BitwiseShiftRightAssign,
		BitwiseXorAssign,
		DivideAssign,
		MinusAssign,
		PlusAssign,
		PlusPlus,
		MinusMinus,
		TimesAssign,
		Return:
		return false

	default:
		return true
	}
}
