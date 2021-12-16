package compiler

import (
	"github.com/NicoNex/tau/code"
	"github.com/NicoNex/tau/obj"
)

type Compilable interface {
	Compile(c *Compiler) (int, error)
}

type EmittedInst struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions code.Instructions
	lastInst     EmittedInst
	prevInst     EmittedInst
}

type Compiler struct {
	constants  []obj.Object
	scopes     []CompilationScope
	scopeIndex int
	*SymbolTable
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []obj.Object
}

func New() *Compiler {
	return &Compiler{
		SymbolTable: NewSymbolTable(),
		scopes:      []CompilationScope{{}},
	}
}

func NewWithState(s *SymbolTable, constants []obj.Object) *Compiler {
	return &Compiler{
		SymbolTable: s,
		scopes:      []CompilationScope{{}},
		constants:   constants,
	}
}

func (c *Compiler) AddConstant(o obj.Object) int {
	c.constants = append(c.constants, o)
	return len(c.constants) - 1
}

func (c *Compiler) AddInstruction(ins []byte) int {
	var posNewInstruction = len(c.scopes[c.scopeIndex].instructions)

	c.scopes[c.scopeIndex].instructions = append(c.scopes[c.scopeIndex].instructions, ins...)
	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op code.Opcode, pos int) {
	prev := c.scopes[c.scopeIndex].lastInst
	last := EmittedInst{op, pos}
	c.scopes[c.scopeIndex].prevInst = prev
	c.scopes[c.scopeIndex].lastInst = last
}

func (c *Compiler) Emit(opcode code.Opcode, operands ...int) int {
	ins := code.Make(opcode, operands...)
	pos := c.AddInstruction(ins)
	c.setLastInstruction(opcode, pos)
	return pos
}

func (c *Compiler) LastIs(op code.Opcode) bool {
	if len(c.scopes[c.scopeIndex].instructions) == 0 {
		return false
	}
	return c.scopes[c.scopeIndex].lastInst.Opcode == op
}

func (c *Compiler) RemoveLast() {
	last := c.scopes[c.scopeIndex].lastInst
	prev := c.scopes[c.scopeIndex].prevInst

	old := c.scopes[c.scopeIndex].instructions
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInst = prev
}

func (c *Compiler) replaceInstruction(pos int, newInst []byte) {
	ins := c.scopes[c.scopeIndex].instructions

	for i := 0; i < len(newInst); i++ {
		ins[pos+i] = newInst[i]
	}
}

func (c *Compiler) ReplaceOperand(opPos, operand int) {
	op := code.Opcode(c.scopes[c.scopeIndex].instructions[opPos])
	newInst := code.Make(op, operand)
	c.replaceInstruction(opPos, newInst)
}

func (c *Compiler) ReplaceLastPopWithReturn() {
	lastPos := c.scopes[c.scopeIndex].lastInst.Position
	c.replaceInstruction(lastPos, code.Make(code.OpReturnValue))
	c.scopes[c.scopeIndex].lastInst.Opcode = code.OpReturnValue
}

func (c *Compiler) EnterScope() {
	c.scopes = append(c.scopes, CompilationScope{})
	c.scopeIndex++
	c.SymbolTable = NewEnclosedSymbolTable(c.SymbolTable)
}

func (c *Compiler) LeaveScope() code.Instructions {
	ins := c.scopes[c.scopeIndex].instructions
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.SymbolTable = c.SymbolTable.outer

	return ins
}

// Returns the position to the last instruction.
func (c *Compiler) Pos() int {
	return len(c.scopes[c.scopeIndex].instructions)
}

func (c *Compiler) Compile(node Compilable) error {
	_, err := node.Compile(c)
	return err
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.scopes[c.scopeIndex].instructions,
		Constants:    c.constants,
	}
}
