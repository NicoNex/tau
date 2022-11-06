package main

import (
	"fmt"
	"testing"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/vm"
)

type vmTestCase struct {
	input    string
	expected any
}

func runVmTests(t *testing.T, tests []vmTestCase) {
	t.Helper()
	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		vm := vm.New("<test input>", comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		if g, ok := stackElem.(obj.Getter); ok {
			stackElem = g.Object()
		}
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T, expected any, actual obj.Object) {
	t.Helper()
	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
		}

	case float64:
		err := testFloatObject(float64(expected), actual)
		if err != nil {
			t.Errorf("testFloatObject failed: %s", err)
		}

	case bool:
		err := testBooleanObject(bool(expected), actual)
		if err != nil {
			t.Errorf("testBooleanObject failed: %s", err)
		}

	case *obj.Null:
		if actual != obj.NullObj {
			t.Errorf("object is not Null: %T (%+v)", actual, actual)
		}

	case string:
		err := testCompilerStringObject(expected, actual)
		if err != nil {
			t.Errorf("testStringObject failed: %s", err)
		}

	case []int:
		l, ok := actual.(obj.List)
		list := l.Val()
		if !ok {
			t.Errorf("object not list: %T (%+v)", actual, actual)
			return
		}

		if len(list) != len(expected) {
			t.Errorf("wrong num of elements. want=%d, got=%d", len(expected), len(list))
			return
		}
		for i, expectedElem := range expected {
			err := testIntegerObject(int64(expectedElem), list[i])
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}

	case map[obj.KeyHash]int64:
		m, ok := actual.(obj.Map)
		if !ok {
			t.Errorf("object is not Hash. got=%T (%+v)", actual, actual)
			return
		}
		mapObj := map[obj.KeyHash]obj.MapPair(m)

		if len(mapObj) != len(expected) {
			t.Errorf("hash has wrong number of Pairs. want=%d, got=%d", len(expected), len(mapObj))
			return
		}

		for expectedKey, expectedValue := range expected {
			pair, ok := mapObj[expectedKey]
			if !ok {
				t.Errorf("no pair for given key in Pairs")
			}
			err := testIntegerObject(expectedValue, pair.Value)
			if err != nil {
				t.Errorf("testIntegerObject failed: %s", err)
			}
		}

	case obj.Error:
		errObj, ok := actual.(obj.Error)
		if !ok {
			t.Errorf("object is not Error: %T (%+v)", actual, actual)
			return
		}
		if errObj.String() != expected.String() {
			t.Errorf("wrong error message. expected=%q, got=%q", expected.String(), errObj.String())
		}
	}
}

func testCompilerStringObject(expected string, actual obj.Object) error {
	result, ok := actual.(obj.String)
	if !ok {
		return fmt.Errorf("object is not string. got=%T (%+v)", actual, actual)
	}
	if result.Val() != expected {
		return fmt.Errorf("object has wrong value. got=%q, want=%q", result.Val(), expected)
	}
	return nil
}

func testBooleanObject(expected bool, actual obj.Object) error {
	result, ok := actual.(*obj.Boolean)
	if !ok {
		return fmt.Errorf("object is not Boolean. got=%T (%+v)", actual, actual)
	}
	if result.Val() != expected {
		return fmt.Errorf("object has wrong value. got=%t, want=%t", result.Val(), expected)
	}
	return nil
}

func TestVMIntegerArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1", 1},
		{"2", 2},
		{"1 + 2", 3},
		{"1 - 2", -1},
		{"1 * 2", 2},
		{"4 / 2", 2.0},
		{"50 / 2 * 2 + 10 - 5", 55.0},
		{"5 * (2 + 10)", 60},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50.0},
	}

	runVmTests(t, tests)
}

func TestVMFloatArithmetic(t *testing.T) {
	tests := []vmTestCase{
		{"1.5", 1.5},
		{"2.8", 2.8},
		{"1.5 + 2.7", 4.2},
		{"1.5 - 2", -0.5},
		{"1.5 * 2", 3.0},
		{"5.0 / 2", 2.5},
	}

	runVmTests(t, tests)
}

func TestVMBooleanExpression(t *testing.T) {
	tests := []vmTestCase{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"true && false", false},
		{"true || false", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
		{"1 <= 1", true},
		{"1 >= 1", true},
		{"2 <= 1", false},
		{"2 >= 1", true},
		{"!(if false { 5; })", true},
	}

	runVmTests(t, tests)
}

func TestVMConditionals(t *testing.T) {
	tests := []vmTestCase{
		{"if true { 10 }", 10},
		{"if true { 10 } else { 20 }", 10},
		{"if false { 10 } else { 20 } ", 20},
		{"if 1 { 10 }", 10},
		{"if 1 < 2 { 10 }", 10},
		{"if 1 < 2 { 10 } else { 20 }", 10},
		{"if 1 > 2 { 10 } else { 20 }", 20},
		{"if 1 > 2 { 10 }", obj.NullObj},
		{"if false { 10 }", obj.NullObj},
		{"if (if false { 10 }) { 10 } else { 20 }", 20},
	}

	runVmTests(t, tests)
}

func TestVMGlobalAssignments(t *testing.T) {
	tests := []vmTestCase{
		{"one = 1; one", 1},
		{"one = 1; two = 2; one + two", 3},
		{"one = 1; two = one + one; one + two", 3},
	}

	runVmTests(t, tests)
}

func TestVMStringExpressions(t *testing.T) {
	tests := []vmTestCase{
		{`"tau"`, "tau"},
		{`"tau" + "rocks"`, "taurocks"},
		{`"t" + "a" + "u"`, "tau"},
	}

	runVmTests(t, tests)
}

func TestVMStringInterpolation(t *testing.T) {
	tests := []vmTestCase{
		{`a = 1; "{if a > 0 { \"test1\" } else { \"test0\" }}"`, "test1"},
		{`"{\"}}\"}"`, "}"},
		{`"{ \"{{\" }"`, "{"},
		{`"{ {\"test\": \"it works\"}[\"test\"] }"`, "it works"},
	}

	runVmTests(t, tests)
}

func TestVMListLiterals(t *testing.T) {
	tests := []vmTestCase{
		{"[]", []int{}},
		{"[1, 2, 3]", []int{1, 2, 3}},
		{"[1 + 2, 3 * 4, 5 + 6]", []int{3, 12, 11}},
	}

	runVmTests(t, tests)
}

func TestVMMapLiterals(t *testing.T) {
	tests := []vmTestCase{
		{
			"{}", map[obj.KeyHash]int64{},
		},
		{
			"{1: 2, 2: 3}",
			map[obj.KeyHash]int64{
				obj.NewInteger(1).(obj.Integer).KeyHash(): 2,
				obj.NewInteger(2).(obj.Integer).KeyHash(): 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[obj.KeyHash]int64{
				obj.NewInteger(2).(obj.Integer).KeyHash(): 4,
				obj.NewInteger(6).(obj.Integer).KeyHash(): 16,
			},
		},
	}
	runVmTests(t, tests)
}

func TestVMIndexExpressions(t *testing.T) {
	tests := []vmTestCase{
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][0 + 2]", 3},
		{"[[1, 1, 1]][0][0]", 1},
		{"{1: 1, 2: 2}[1]", 1},
		{"{1: 1, 2: 2}[2]", 2},
	}

	runVmTests(t, tests)
}

func TestVMCallingFunctionsWithoutArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
fivePlusTen = fn() { 5 + 10 }
fivePlusTen()
`,
			expected: 15,
		},
		{
			input: `
one = fn() { 1 }
two = fn() { 2 }
one() + two()
`,
			expected: 3,
		},
		{
			input: `
a = fn() { 1 }
b = fn() { a() + 1 }
c = fn() { b() + 1 }
c()
`,
			expected: 3,
		},
	}

	runVmTests(t, tests)
}

func TestVMFunctionsWithReturnStatement(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
earlyExit = fn() { return 99; 100; }
earlyExit()
`,
			expected: 99,
		},
		{
			input: `
earlyExit = fn() { return 99; return 100; }
earlyExit()
`,
			expected: 99,
		},
		{
			input: `
noReturn = fn() { }
noReturn()
`,
			expected: obj.NullObj,
		},
		{
			input: `
noReturn = fn() { }
noReturnTwo = fn() { noReturn() }
noReturn()
noReturnTwo()
`,
			expected: obj.NullObj,
		},
	}
	runVmTests(t, tests)
}

func TestVMFirstClassFunctions(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
returnsOne = fn() { 1 }
returnsOneReturner = fn() { returnsOne }
returnsOneReturner()()
`,
			expected: 1,
		},
		{
			input: `
returnsOneReturner = fn() {
	returnsOne = fn() { 1 }
	returnsOne
};
returnsOneReturner()();
`,
			expected: 1,
		},
	}

	runVmTests(t, tests)
}

func TestVMCallingFunctionsWithBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
one = fn() { one = 1; one };
one();
`,
			expected: 1,
		},
		{
			input: `
oneAndTwo = fn() { one = 1; two = 2; one + two; };
oneAndTwo();
`,
			expected: 3,
		}, {
			input: `
oneAndTwo = fn() { one = 1; two = 2; one + two; };
threeAndFour = fn() { three = 3; four = 4; three + four; };
oneAndTwo() + threeAndFour();
`,
			expected: 10,
		},
		{
			input: `
firstFoobar = fn() { foobar = 50; foobar; };
secondFoobar = fn() { foobar = 100; foobar; };
firstFoobar() + secondFoobar();
`,
			expected: 150,
		},
		{
			input: `
globalSeed = 50;
minusOne = fn() {
num = 1;
globalSeed - num;
}
minusTwo = fn() {
num = 2;
globalSeed - num;
}
minusOne() + minusTwo();
`,
			expected: 97,
		},
	}
	runVmTests(t, tests)
}

func TestVMCallingFunctionsWithArgumentsAndBindings(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
identity = fn(a) { a; }
identity(4)
`,
			expected: 4,
		},
		{
			input: `
sum = fn(a, b) { a + b; }
sum(1, 2)
`,
			expected: 3,
		},
		{
			input: `
sum = fn(a, b) {
	c = a + b
	c
}
sum(1, 2)
`,
			expected: 3,
		},
		{
			input: `
sum = fn(a, b) {
	c = a + b
	c
}
sum(1, 2) + sum(3, 4)`,
			expected: 10,
		},
		{
			input: `
sum = fn(a, b) {
	c = a + b
	c
}
outer = fn() {
	sum(1, 2) + sum(3, 4)
}
outer()
`,
			expected: 10,
		},
		{
			input: `
globalNum = 10

sum = fn(a, b) {
	c = a + b
	c + globalNum
}

outer = fn() {
	sum(1, 2) + sum(3, 4) + globalNum
}

outer() + globalNum
`,
			expected: 50,
		},
	}

	runVmTests(t, tests)
}

func TestCallingFunctionsWithWrongArguments(t *testing.T) {
	tests := []vmTestCase{
		{
			input:    `fn() { 1 }(1);`,
			expected: `wrong number of arguments: expected 0, got 1`,
		},
		{
			input:    `fn(a) { a }();`,
			expected: `wrong number of arguments: expected 1, got 0`,
		},
		{
			input:    `fn(a, b) { a + b; }(1);`,
			expected: `wrong number of arguments: expected 2, got 1`,
		},
	}

	for _, tt := range tests {
		program := parse(tt.input)
		comp := compiler.New()
		err := comp.Compile(program)
		if err != nil {
			t.Fatalf("compiler error: %s", err)
		}
		vm := vm.New("<test input>", comp.Bytecode())
		err = vm.Run()
		if err == nil {
			t.Fatalf("expected VM error but resulted in none.")
		}
		if err.Error() != tt.expected {
			t.Fatalf("wrong VM error: want=%q, got=%q", tt.expected, err)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []vmTestCase{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		// {
		// 	`len(1)`,
		// 	&object.Error{
		// 		Message: "argument to `len` not supported, got INTEGER",
		// 	},
		// },
		// {`len("one", "two")`,
		// 	&object.Error{
		// 		Message: "wrong number of arguments. got=2, want=1",
		// 	},
		// },
		{`len([1, 2, 3])`, 3},
		{`len([])`, 0},
		{`println("hello", "world!")`, obj.NullObj},
		// {`first([1, 2, 3])`, 1},
		// {`first([])`, Null},
		// {`first(1)`,
		// 	&object.Error{
		// 		Message: "argument to `first` must be ARRAY, got INTEGER",
		// 	},
		// },
		// {`last([1, 2, 3])`, 3},
		// {`last([])`, Null},
		// {`last(1)`,
		// 	&object.Error{
		// 		Message: "argument to `last` must be ARRAY, got INTEGER",
		// 	},
		// },
		// {`rest([1, 2, 3])`, []int{2, 3}},
		// {`rest([])`, Null},
		// {`push([], 1)`, []int{1}},
		// {`push(1, 1)`, &object.Error{
		// 	Message: "argument to `push` must be ARRAY, got INTEGER",
		// },
		// },
	}

	runVmTests(t, tests)
}

func TestVMClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
newClosure = fn(a) {
	fn() { a }
}
closure = newClosure(99)
closure()
`,
			expected: 99,
		},
		{
			input: `
newAdder = fn(a, b) {
	fn(c) { a + b + c }
};
adder = newAdder(1, 2)
adder(8)
`,
			expected: 11,
		},
		{
			input: `
newAdder = fn(a, b) {
	c = a + b
	fn(d) { c + d }
};
adder = newAdder(1, 2)
adder(8)
`,
			expected: 11,
		},
		{
			input: `
newAdderOuter = fn(a, b) {
	c = a + b
	fn(d) {
		e = d + c
		fn(f) { e + f }
	}
}
newAdderInner = newAdderOuter(1, 2)
adder = newAdderInner(3)
adder(8)
`,
			expected: 14,
		},
		{
			input: `
a = 1
newAdderOuter = fn(b) {
	fn(c) {
		fn(d) { a + b + c + d }
	}
}
newAdderInner = newAdderOuter(2)
adder = newAdderInner(3)
adder(8)
`,
			expected: 14,
		},
		{
			input: `
newClosure = fn(a, b) {
	one = fn() { a }
	two = fn() { b }
	fn() { one() + two() }
}
closure = newClosure(9, 90)
closure()
`,
			expected: 99,
		},
	}
	runVmTests(t, tests)
}

func TestVMRecursiveClosures(t *testing.T) {
	tests := []vmTestCase{
		{
			input: `
countDown = fn(x) {
	if x == 0 {
		return 0;
	}
	countDown(x-1)
}

countDown(3)
`,
			expected: 0,
		},
	}

	runVmTests(t, tests)
}
