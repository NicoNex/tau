package main

import (
	"fmt"
	"testing"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
)

type compilerTestCase struct {
	input                string
	expectedConstants    []any
	expectedInstructions []code.Instructions
}

func parse(input string) compiler.Compilable {
	res, _ := parser.Parse(input)
	return res
}

func concatInstructions(s []code.Instructions) code.Instructions {
	var out = code.Instructions{}

	for _, ins := range s {
		out = append(out, ins...)
	}
	return out
}

func testInstructions(expected []code.Instructions, actual code.Instructions) error {
	concatted := concatInstructions(expected)

	if len(actual) != len(concatted) {
		return fmt.Errorf("wrong instructions length.\nwant=%q\ngot =%q", concatted, actual)
	}
	for i, ins := range concatted {
		if actual[i] != ins {
			return fmt.Errorf("wrong instruction at %d.\nwant=%q\ngot =%q", i, concatted, actual)
		}
	}
	return nil
}

func testStringObject(expected string, actual obj.Object) error {
	result, ok := actual.(obj.String)
	if !ok {
		return fmt.Errorf("object is not string. got=%T (%+v)", actual, actual)
	}
	if result.Val() != expected {
		return fmt.Errorf("object has wrong value. got=%q, want=%q", result.Val(), expected)
	}
	return nil
}

func TestCompilerIntegerArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 + 2",
			expectedConstants: []any{3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 - 2",
			expectedConstants: []any{-1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "2 * 2",
			expectedConstants: []any{4},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "4 / 2",
			expectedConstants: []any{2.0},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "4 % 2",
			expectedConstants: []any{0},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1; 2",
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "-1",
			expectedConstants: []any{-1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop)},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerFloatArithmetic(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1.5; 2.3",
			expectedConstants: []any{1.5, 2.3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1.5 + 2.5",
			expectedConstants: []any{4.0},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1.5 - 2",
			expectedConstants: []any{-0.5},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1.5 * 2",
			expectedConstants: []any{3.0},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "2.8 / 2",
			expectedConstants: []any{1.4},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "-1.37",
			expectedConstants: []any{-1.37},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop)},
		},
	}

	runCompilerTests(t, tests)
}

func runCompilerTests(t *testing.T, tests []compilerTestCase) {
	t.Helper()
	for _, tt := range tests {
		program := parse(tt.input)
		compiler := compiler.New()

		err := compiler.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}

		bytecode := compiler.Bytecode()
		err = testInstructions(tt.expectedInstructions, bytecode.Instructions)
		if err != nil {
			t.Fatalf("testInstructions failed: %s", err)
		}

		err = testConstants(t, tt.expectedConstants, bytecode.Constants)
		if err != nil {
			t.Fatalf("testConstants failed: %s", err)
		}
	}
}

func testConstants(t *testing.T, expected []any, actual []obj.Object) error {
	if len(expected) != len(actual) {
		return fmt.Errorf("wrong number of constants. got=%d, want=%d",
			len(actual), len(expected))
	}
	for i, constant := range expected {
		switch constant := constant.(type) {
		case int:
			err := testIntegerObject(int64(constant), actual[i])
			if err != nil {
				return fmt.Errorf("constant %d - testIntegerObject failed: %s",
					i, err)
			}

		case string:
			if err := testStringObject(constant, actual[i]); err != nil {
				return fmt.Errorf("constant %d - not a function: %T", i, err)
			}

		case []code.Instructions:
			fn, ok := actual[i].(obj.Function)
			if !ok {
				return fmt.Errorf("constant %d - testInstructions failed: %s", i, actual[i])
			}

			err := testInstructions(constant, fn.Instructions)
			if err != nil {
				return fmt.Errorf("constant %d - testInstructions failed: %s", i, err)
			}
		}
	}
	return nil
}

func testIntegerObject(expected int64, actual obj.Object) error {
	result, ok := actual.(obj.Integer)
	if !ok {
		return fmt.Errorf("object is not integer. got=%T (%+v)", actual, actual)
	}
	if result.Val() != expected {
		return fmt.Errorf("object has wrong value. got=%d, want=%d", result.Val(), expected)
	}
	return nil
}

func testFloatObject(expected float64, actual obj.Object) error {
	result, ok := actual.(obj.Float)
	if !ok {
		return fmt.Errorf("object is not float. got=%T (%+v)", actual, actual)
	}
	if result.Val() != expected {
		return fmt.Errorf("object has wrong value. got=%f, want=%f", result.Val(), expected)
	}
	return nil
}

func TestCompilerBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "true",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpTrue),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "false",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpFalse),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestBooleanExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "1 > 2",
			expectedConstants: []any{false},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 < 2",
			expectedConstants: []any{true},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 >= 2",
			expectedConstants: []any{false},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 <= 2",
			expectedConstants: []any{true},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 == 2",
			expectedConstants: []any{false},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "1 != 2",
			expectedConstants: []any{true},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true == false",
			expectedConstants: []any{false},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true != false",
			expectedConstants: []any{true},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true && false",
			expectedConstants: []any{false},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "true || false",
			expectedConstants: []any{true},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "!true",
			expectedConstants: []any{false},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestCompilerConditionals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `if true { 10 }; 3333;`,
			expectedConstants: []any{10, 3333},
			expectedInstructions: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 10),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 11),
				// 0010
				code.Make(code.OpNull),
				// 0011
				code.Make(code.OpPop),
				// 0012
				code.Make(code.OpConstant, 1),
				// 0015
				code.Make(code.OpPop),
			},
		},
		{
			input:             `if true { 10 } else { 20 }; 3333;`,
			expectedConstants: []any{10, 20, 3333},
			expectedInstructions: []code.Instructions{
				// 0000
				code.Make(code.OpTrue),
				// 0001
				code.Make(code.OpJumpNotTruthy, 10),
				// 0004
				code.Make(code.OpConstant, 0),
				// 0007
				code.Make(code.OpJump, 13),
				// 0010
				code.Make(code.OpConstant, 1),
				// 0013
				code.Make(code.OpPop),
				// 0014
				code.Make(code.OpConstant, 2),
				// 0017
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestGlobalAssignments(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `one = 1; two = 2;`,
			expectedConstants: []any{1, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `one = 1; one;`,
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `one = 1; two = one; two;`,
			expectedConstants: []any{1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpSetGlobal, 1),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `three = 1 + 2`,
			expectedConstants: []any{3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `three = 4; three = three - 1`,
			expectedConstants: []any{4, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpSub),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `three = 2; three += 1`,
			expectedConstants: []any{2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpAdd),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerStringExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `"tau"`,
			expectedConstants: []any{"tau"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             `"tau" + "rocks"`,
			expectedConstants: []any{"taurocks"},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestCompilerListLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[]",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpList, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1, 2, 3]",
			expectedConstants: []any{1, 2, 3},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpList, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "[1 + 2, 3 - 4, 5 * 6]",
			expectedConstants: []any{3, -1, 30},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpList, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerMapLiterals(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "{}",
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpMap, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2, 3: 4, 5: 6}",
			expectedConstants: []any{1, 2, 3, 4, 5, 6},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpConstant, 4),
				code.Make(code.OpConstant, 5),
				code.Make(code.OpMap, 6),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2 + 3, 4: 5 * 6}",
			expectedConstants: []any{1, 5, 4, 30},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpMap, 4),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerIndexExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             "[1, 2, 3][1 + 1]",
			expectedConstants: []any{1, 2, 3, 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpList, 3),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
		{
			input:             "{1: 2}[2 - 1]",
			expectedConstants: []any{1, 2, 1},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpMap, 2),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpIndex),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerDotExpressions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
a = new();
a.a = 2`,
			expectedConstants: []any{"a", 2},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetBuiltin, 16),
				code.Make(code.OpCall, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 0),
				code.Make(code.OpDot),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpDefine),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerFunctions(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { return 5 + 10 }`,
			expectedConstants: []any{
				15,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { 1; 2 }`,
			expectedConstants: []any{
				1,
				2,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestFunctionsWithoutReturnValue(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { }`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestCompilerFunctionCalls(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `fn() { 24 }()`,
			expectedConstants: []any{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0), // The literal "24"
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0), // The compiled function
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
noArg = fn() { 24 }
noArg()
`,
			expectedConstants: []any{
				24,
				[]code.Instructions{
					code.Make(code.OpConstant, 0), // The literal "24"
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0), // The compiled function
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpCall, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
oneArg = fn(a) { }
oneArg(24)
`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
				24,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
manyArg = fn(a, b, c) { }
manyArg(24, 25, 26)
`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpReturn),
				},
				24,
				25,
				26,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
oneArg = fn(a) { a }
oneArg(24)
`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
				24,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
manyArg = fn(a, b, c) { a; b; c }
manyArg(24, 25, 26)
`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 2),
					code.Make(code.OpReturnValue),
				},
				24,
				25,
				26,
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpGetGlobal, 0),
				code.Make(code.OpConstant, 1),
				code.Make(code.OpConstant, 2),
				code.Make(code.OpConstant, 3),
				code.Make(code.OpCall, 3),
				code.Make(code.OpPop),
			},
		},
	}

	runCompilerTests(t, tests)
}

func TestCompilerAssignStatementScopes(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
num = 55
fn() { num }
`,
			expectedConstants: []any{
				55,
				[]code.Instructions{
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
fn() {
	num = 55
	num
}
`,
			expectedConstants: []any{
				55,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
fn() {
	a = 55;
	b = 77;
	a + b
}
`,
			expectedConstants: []any{
				55,
				77,
				[]code.Instructions{
					code.Make(code.OpConstant, 0),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 1),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpGetLocal, 1),
					code.Make(code.OpAdd), code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestCompilerBuiltins(t *testing.T) {
	tests := []compilerTestCase{
		{
			input:             `len([])`,
			expectedConstants: []any{},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpGetBuiltin, 0),
				code.Make(code.OpList, 0),
				code.Make(code.OpCall, 1),
				code.Make(code.OpPop),
			},
		},
		{
			input: `fn() { len([]) }`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetBuiltin, 0),
					code.Make(code.OpList, 0),
					code.Make(code.OpCall, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 0, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}

func TestClosures(t *testing.T) {
	tests := []compilerTestCase{
		{
			input: `
fn(a) {
	fn(b) {
		a + b
	}
}
`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				}, []code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 1, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
fn(a) {
	fn(b) {
		fn(c) {
			a + b + c
		}
	}
}
`,
			expectedConstants: []any{
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetFree, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 0, 2),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 1, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpClosure, 2, 0),
				code.Make(code.OpPop),
			},
		},
		{
			input: `
global = 55
fn() {
	a = 66
	fn() {
		b = 77
		fn() {
			c = 88
			global + a + b + c
		}
	}
}
`,
			expectedConstants: []any{
				55,
				66,
				77,
				88,
				[]code.Instructions{
					code.Make(code.OpConstant, 3),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetGlobal, 0),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpGetFree, 1),
					code.Make(code.OpAdd),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpAdd),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 2),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetFree, 0),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 4, 2),
					code.Make(code.OpReturnValue),
				},
				[]code.Instructions{
					code.Make(code.OpConstant, 1),
					code.Make(code.OpSetLocal, 0),
					code.Make(code.OpPop),
					code.Make(code.OpGetLocal, 0),
					code.Make(code.OpClosure, 5, 1),
					code.Make(code.OpReturnValue),
				},
			},
			expectedInstructions: []code.Instructions{
				code.Make(code.OpConstant, 0),
				code.Make(code.OpSetGlobal, 0),
				code.Make(code.OpPop),
				code.Make(code.OpClosure, 6, 0),
				code.Make(code.OpPop),
			},
		},
	}
	runCompilerTests(t, tests)
}
