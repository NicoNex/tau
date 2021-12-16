package compiler

import (
	"testing"

	"github.com/NicoNex/tau/code"
)

func TestCompilerScopes(t *testing.T) {
	compiler := New()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 0)
	}
	globalSymbolTable := compiler.SymbolTable
	compiler.Emit(code.OpMul)
	compiler.EnterScope()
	if compiler.scopeIndex != 1 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d", compiler.scopeIndex, 1)
	}
	compiler.Emit(code.OpSub)
	if len(compiler.scopes[compiler.scopeIndex].instructions) != 1 {
		t.Errorf("instructions length wrong. got=%d",
			len(compiler.scopes[compiler.scopeIndex].instructions))
	}
	last := compiler.scopes[compiler.scopeIndex].lastInst
	if last.Opcode != code.OpSub {
		t.Errorf("lastInst.Opcode wrong. got=%d, want=%d",
			last.Opcode, code.OpSub)
	}
	if compiler.SymbolTable.outer != globalSymbolTable {
		t.Errorf("compiler did not enclose symbolTable")
	}
	compiler.LeaveScope()
	if compiler.scopeIndex != 0 {
		t.Errorf("scopeIndex wrong. got=%d, want=%d",
			compiler.scopeIndex, 0)
	}
	if compiler.SymbolTable != globalSymbolTable {
		t.Errorf("compiler did not restore global symbol table")
	}
	if compiler.SymbolTable.outer != nil {
		t.Errorf("compiler modified global symbol table incorrectly")
	}
	compiler.Emit(code.OpAdd)
	if len(compiler.scopes[compiler.scopeIndex].instructions) != 2 {
		t.Errorf("instructions length wrong. got=%d",
			len(compiler.scopes[compiler.scopeIndex].instructions))
	}
	last = compiler.scopes[compiler.scopeIndex].lastInst
	if last.Opcode != code.OpAdd {
		t.Errorf("lastInst.Opcode wrong. got=%d, want=%d",
			last.Opcode, code.OpAdd)
	}
	previous := compiler.scopes[compiler.scopeIndex].prevInst
	if previous.Opcode != code.OpMul {
		t.Errorf("prevInst.Opcode wrong. got=%d, want=%d",
			previous.Opcode, code.OpMul)
	}
}
