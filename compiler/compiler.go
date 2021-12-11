package compiler

import (
	// "github.com/NicoNex/tau/ast"
	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/obj"
)

type Compiler struct {
	instructions code.Instructions
	constants    []obj.Object
}

type Compilable interface {
	Compile(c *Compiler) int
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []obj.Object
}

func New() *Compiler {
	return &Compiler{
		instructions: code.Instructions{},
		constants:    []obj.Object{},
	}
}

func (c *Compiler) AddConstant(o obj.Object) int {
	c.constants = append(c.constants, o)
	return len(c.constants) - 1
}

func (c *Compiler) AddInstruction(ins []byte) int {
	var posNewInstruction = len(c.instructions)
	c.instructions = append(c.instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) Emit(opcode code.Opcode, operands ...int) int {
	ins := code.Make(opcode, operands...)
	return c.AddInstruction(ins)
}

func (c *Compiler) Compile(node Compilable) error {
	node.Compile(c)
	return nil
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.instructions,
		Constants:    c.constants,
	}
}
