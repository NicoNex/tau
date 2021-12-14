package main

import (
	"fmt"
	"testing"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
	"github.com/NicoNex/tau/vm"
)

type vmTestCase struct {
	input    string
	expected interface{}
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
		vm := vm.New(comp.Bytecode())
		err = vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}
		stackElem := vm.LastPoppedStackElem()
		testExpectedObject(t, tt.expected, stackElem)
	}
}

func testExpectedObject(t *testing.T, expected interface{}, actual obj.Object) {
	t.Helper()
	switch expected := expected.(type) {
	case int:
		err := testIntegerObject(int64(expected), actual)
		if err != nil {
			t.Errorf("testIntegerObject failed: %s", err)
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
	}
}

func testCompilerStringObject(expected string, actual obj.Object) error {
	result, ok := actual.(*obj.String)
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
		{"4 / 2", 2},
		{"50 / 2 * 2 + 10 - 5", 55},
		{"5 * (2 + 10)", 60},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"5 * (2 + 10)", 60},
		{"-5", -5},
		{"-10", -10},
		{"-50 + 100 + -50", 0},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
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
				obj.NewInteger(1).(*obj.Integer).KeyHash(): 2,
				obj.NewInteger(2).(*obj.Integer).KeyHash(): 3,
			},
		},
		{
			"{1 + 1: 2 * 2, 3 + 3: 4 * 4}",
			map[obj.KeyHash]int64{
				obj.NewInteger(2).(*obj.Integer).KeyHash(): 4,
				obj.NewInteger(6).(*obj.Integer).KeyHash(): 16,
			},
		},
	}
	runVmTests(t, tests)
}
