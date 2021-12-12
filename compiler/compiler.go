package compiler

import (
	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/obj"
)

type Compilable interface {
	Compile(c *Compiler) int
}

type EmittedInst struct {
	Opcode   code.Opcode
	Position int
}

type Compiler struct {
	instructions code.Instructions
	constants    []obj.Object
	lastInst     EmittedInst
	prevInst     EmittedInst
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

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	prev := c.lastInst
	last := EmittedInst{op, pos}
	c.prevInst = prev
	c.lastInst = last
}

func (c *Compiler) Emit(opcode code.Opcode, operands ...int) int {
	ins := code.Make(opcode, operands...)
	pos := c.AddInstruction(ins)
	c.setLastInstruction(opcode, pos)
	return pos
}

func (c *Compiler) LastIsPop() bool {
	return c.lastInst.Opcode == code.OpPop
}

func (c *Compiler) RemoveLast() {
	c.instructions = c.instructions[:c.lastInst.Position]
	c.lastInst = c.prevInst
}

func (c *Compiler) replaceInstruction(pos int, newInst []byte) {
	for i := 0; i < len(newInst); i++ {
		c.instructions[pos+i] = newInst[i]
	}
}

func (c *Compiler) ReplaceOperand(opPos, operand int) {
	op := code.Opcode(c.instructions[opPos])
	newInst := code.Make(op, operand)
	c.replaceInstruction(opPos, newInst)
}

// Returns the position to the last instruction.
func (c *Compiler) Pos() int {
	return len(c.instructions)
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
