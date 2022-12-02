package compiler

import (
	"errors"
	"fmt"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/tauerr"
)

type Compilable interface {
	Compile(c *Compiler) (int, error)
	IsConstExpression() bool
}

type EmittedInst struct {
	Opcode   code.Opcode
	Position int
}

type CompilationScope struct {
	instructions code.Instructions
	lastInst     EmittedInst
	prevInst     EmittedInst
	bookmarks    []tauerr.Bookmark
}

type Compiler struct {
	constants   *[]obj.Object
	scopes      []CompilationScope
	scopeIndex  int
	fileName    string
	fileContent string
	*SymbolTable
}

type Bytecode struct {
	Instructions code.Instructions
	Constants    []obj.Object
	Bookmarks    []tauerr.Bookmark
}

const (
	ContinuePlaceholder = 9998
	BreakPlaceholder    = 9997
)

func New() *Compiler {
	var st = NewSymbolTable()

	for i, b := range obj.Builtins {
		st.DefineBuiltin(i, b.Name)
	}

	return &Compiler{
		SymbolTable: st,
		scopes:      []CompilationScope{{}},
		constants:   &[]obj.Object{},
	}
}

func NewWithState(s *SymbolTable, constants *[]obj.Object) *Compiler {
	return &Compiler{
		SymbolTable: s,
		scopes:      []CompilationScope{{}},
		constants:   constants,
	}
}

func (c *Compiler) AddConstant(o obj.Object) int {
	*c.constants = append(*c.constants, o)
	return len(*c.constants) - 1
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

func (c *Compiler) ReplaceContinueOperands(startBody, endBody, operand int) error {
	ins := c.scopes[c.scopeIndex].instructions
	l := len(ins)

	if startBody > l || endBody > l {
		return errors.New("compiler error: startBody or endBody positions out of range")
	}

	for i := startBody; i < endBody && i < l; {
		def, err := code.Lookup(ins[i])
		if err != nil {
			return err
		}

		operands, read := code.ReadOperands(def, ins[i+1:])
		opcode := code.Opcode(ins[i])

		if opcode == code.OpJump && operands[0] == ContinuePlaceholder {
			c.ReplaceOperand(i, operand)
		}

		i += read + 1
	}
	return nil
}

func (c *Compiler) ReplaceBreakOperands(startBody, endBody, operand int) error {
	ins := c.scopes[c.scopeIndex].instructions
	l := len(ins)

	if startBody > l || endBody > l {
		return errors.New("compiler error: startBody or endBody positions out of range")
	}

	for i := startBody; i < endBody && i < l; {
		def, err := code.Lookup(ins[i])
		if err != nil {
			return err
		}

		operands, read := code.ReadOperands(def, ins[i+1:])
		opcode := code.Opcode(ins[i])

		if opcode == code.OpJump && operands[0] == BreakPlaceholder {
			c.ReplaceOperand(i, operand)
		}

		i += read + 1
	}
	return nil
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

func (c *Compiler) LeaveScope() (code.Instructions, []tauerr.Bookmark) {
	ins := c.scopes[c.scopeIndex].instructions
	bookmarks := c.scopes[c.scopeIndex].bookmarks
	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--
	c.SymbolTable = c.SymbolTable.outer

	return ins, bookmarks
}

// Returns the position to the last instruction.
func (c *Compiler) Pos() int {
	return len(c.scopes[c.scopeIndex].instructions)
}

func (c *Compiler) Bookmark(pos int) {
	if c.fileContent == "" {
		return
	}

	b := tauerr.NewBookmark(c.fileContent, pos, c.Pos())
	c.scopes[c.scopeIndex].bookmarks = append(c.scopes[c.scopeIndex].bookmarks, b)
}

func (c *Compiler) UnresolvedError(name string, pos int) error {
	if c.fileName == "" || c.fileContent == "" {
		return fmt.Errorf("undefined variable %s", name)
	}

	return tauerr.New(c.fileName, c.fileContent, pos, "undefined variable %s", name)
}

func (c *Compiler) NewError(pos int, s string, a ...any) error {
	if c.fileName == "" || c.fileContent == "" {
		return fmt.Errorf(s, a...)
	}

	return tauerr.New(c.fileName, c.fileContent, pos, s, a...)
}

func (c *Compiler) Compile(node Compilable) error {
	_, err := node.Compile(c)
	return err
}

func (c *Compiler) Bytecode() *Bytecode {
	return &Bytecode{
		Instructions: c.scopes[c.scopeIndex].instructions,
		Constants:    *c.constants,
		Bookmarks:    c.scopes[c.scopeIndex].bookmarks,
	}
}

func (c *Compiler) SetBytecode(b *Bytecode) {
	c.scopes[c.scopeIndex].instructions = b.Instructions
	*c.constants = b.Constants
}

func (c *Compiler) SetFileInfo(name, content string) {
	c.fileName = name
	c.fileContent = content
}

func (c *Compiler) LoadSymbol(s Symbol) int {
	switch s.Scope {
	case GlobalScope:
		return c.Emit(code.OpGetGlobal, s.Index)
	case LocalScope:
		return c.Emit(code.OpGetLocal, s.Index)
	case BuiltinScope:
		return c.Emit(code.OpGetBuiltin, s.Index)
	case FreeScope:
		return c.Emit(code.OpGetFree, s.Index)
	case FunctionScope:
		return c.Emit(code.OpCurrentClosure)
	default:
		return 0
	}
}
