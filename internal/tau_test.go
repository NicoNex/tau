package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

type TauTest map[string]func(o obj.Object) error

func (tt TauTest) add(code string, expected obj.Object) TauTest {
	tt[code] = func(got obj.Object) error {
		if expected.Type() != got.Type() {
			return fmt.Errorf("type mismatch\n%s\nexpected %v, got %v", code, expected.Type(), got.Type())
		}

		switch e := expected.(type) {
		case obj.Null:
			return nil

		case obj.Boolean:
			g := got.(obj.Boolean)
			if e != g {
				return fmt.Errorf("value mismatch\n%s\nexpected %v, got %v", code, expected, got)
			}

		case obj.Integer:
			g := got.(obj.Integer)
			if e != g {
				return fmt.Errorf("value mismatch\n%s\nexpected %v, got %v", code, expected, got)
			}

		case obj.Float:
			g := got.(obj.Float)
			if e != g {
				return fmt.Errorf("value mismatch\n%s\nexpected %v, got %v", code, expected, got)
			}

		case obj.String:
			g := got.(obj.String)
			if e != g {
				return fmt.Errorf("value mismatch\n%s\nexpected %v, got %v", code, expected, got)
			}

		case obj.Error:
			g := got.(obj.Error)
			if e != g {
				return fmt.Errorf("value mismatch\n%s\nexpected %v, got %v", code, expected, got)
			}
		}
		return nil
	}
	return tt
}

func (tt TauTest) run(t *testing.T) {
	for code, fn := range tt {
		bcode, err := compile(code)
		if err != nil {
			t.Log(code)
			t.Fatal(err)
		}
		tvm := vm.New("<tautest>", bcode)
		tvm.Run()
		if err := fn(tvm.LastPoppedStackObj()); err != nil {
			t.Fatal(err)
		}
	}
}

func compile(code string) (bc *compiler.Bytecode, err error) {
	tree, errs := parser.Parse("<tautest>", code)
	if len(errs) > 0 {
		var buf strings.Builder

		buf.WriteString("parser errors:")
		for _, e := range errs {
			buf.WriteString(e.Error())
			buf.WriteByte('\n')
		}
		return nil, errors.New(buf.String())
	}

	c := compiler.New()
	c.SetFileInfo("<tautest>", code)
	if err = c.Compile(tree); err != nil {
		return
	}
	return c.Bytecode(), nil
}

func TestTau(t *testing.T) {
	var tt = make(TauTest)

	// Here we assign values to variables to avoid the constant folding
	// compiler optimization, this test will need to be changed once more
	// optimizations will be added to the compiler or AST.

	// Test assign
	tt.add(`a = 123; a`, obj.NewInteger(123))

	// Test plus
	tt.add(`a = 1; b = 2; a + b`, obj.NewInteger(3))
	tt.add(`a = 1; b = 2; a += b; a`, obj.NewInteger(3))
	tt.add(`a = 1.3; b = 2; a += b; a`, obj.NewFloat(3.3))

	// // Test plusplus
	tt.add(`a = 1; ++a`, obj.NewInteger(2))

	// Test minus
	tt.add(`a = 123; -a`, obj.NewInteger(-123))
	tt.add(`a = 123; b = 3; a - b`, obj.NewInteger(120))
	tt.add(`a = 12.5; b = 2; a - b`, obj.NewFloat(10.5))
	tt.add(`a = 12.5; b = 2; a -= b; a`, obj.NewFloat(10.5))

	// Test minusminus
	tt.add(`a = 1; --a`, obj.NewInteger(0))

	// Test multiply
	tt.add(`a = 2; b = 8; a * b`, obj.NewInteger(16))
	tt.add(`a = 3.5; b = 3; a * b`, obj.NewFloat(10.5))
	tt.add(`a = 2; b = 8; a *= b; a`, obj.NewInteger(16))
	tt.add(`a = 3.5; b = 3; a *= b; a`, obj.NewFloat(10.5))

	// Test divide
	tt.add(`a = 16; b = 2; a / b`, obj.NewFloat(8))
	tt.add(`a = 3; b = 2; a / b`, obj.NewFloat(1.5))
	tt.add(`a = 3; b = 2; a /= b; a`, obj.NewFloat(1.5))

	// Test modulus
	tt.add(`a = 10; b = 3; a % b`, obj.NewInteger(1))
	tt.add(`a = 10; b = 3; a %= b; a`, obj.NewInteger(1))

	// Test logical and
	tt.add(`a = true; b = true; a && b`, obj.ParseBool(true))
	tt.add(`a = true; b = false; a && b`, obj.ParseBool(false))
	tt.add(`a = 123; b = null; a && b`, obj.ParseBool(false))

	// Test logical or
	tt.add(`a = true; b = false; a || b`, obj.ParseBool(true))
	tt.add(`a = false; b = false; a || b`, obj.ParseBool(false))
	tt.add(`a = 123; b = null; a || b`, obj.ParseBool(true))

	// Test bang
	tt.add(`a = 123; !a`, obj.ParseBool(false))
	tt.add(`a = null; !a`, obj.ParseBool(true))

	// Test equality
	tt.add(`a = 123; b = 123; a == b`, obj.ParseBool(true))
	tt.add(`a = 123; b = 456; a == b`, obj.ParseBool(false))
	tt.add(`a = 123; b = null; a == b`, obj.ParseBool(false))

	// Test inequality
	tt.add(`a = 123; b = 456; a != b`, obj.ParseBool(true))
	tt.add(`a = 123; b = 123; a != b`, obj.ParseBool(false))
	tt.add(`a = 123; b = null; a != b`, obj.ParseBool(true))

	// Test less than
	tt.add(`a = 123; b = 456; a < b`, obj.ParseBool(true))
	tt.add(`a = 456; b = 123; a < b`, obj.ParseBool(false))
	tt.add(`a = 123; b = 123; a < b`, obj.ParseBool(false))

	// Test greater than
	tt.add(`a = 456; b = 123; a > b`, obj.ParseBool(true))
	tt.add(`a = 123; b = 456; a > b`, obj.ParseBool(false))
	tt.add(`a = 123; b = 123; a > b`, obj.ParseBool(false))

	// Test less than or equal to
	tt.add(`a = 123; b = 456; a <= b`, obj.ParseBool(true))
	tt.add(`a = 456; b = 123; a <= b`, obj.ParseBool(false))
	tt.add(`a = 123; b = 123; a <= b`, obj.ParseBool(true))

	// Test greater than or equal to
	tt.add(`a = 456; b = 123; a >= b`, obj.ParseBool(true))
	tt.add(`a = 123; b = 456; a >= b`, obj.ParseBool(false))
	tt.add(`a = 123; b = 123; a >= b`, obj.ParseBool(true))

	// Test bitwise not
	tt.add(`a = 2; ~a`, obj.NewInteger(-3))
	tt.add(`a = 2; a = ~a; a`, obj.NewInteger(-3))

	// Test bitwise and
	tt.add(`a = 2; b = 4; a & b`, obj.NewInteger(0))
	tt.add(`a = 2; b = 4; a &= b`, obj.NewInteger(0))

	// Test bitwise or
	tt.add(`a = 3; b = 4; a | b`, obj.NewInteger(7))
	tt.add(`a = 3; b = 4; a |= b; a`, obj.NewInteger(7))

	// Test bitwise xor
	tt.add(`a = 1; b = 8; a ^ b`, obj.NewInteger(9))
	tt.add(`a = 1; b = 8; a ^= b; a`, obj.NewInteger(9))

	// Test bitwise shift left
	tt.add(`a = 1; b = 3; a << b`, obj.NewInteger(8))
	tt.add(`a = 1; b = 3; a <<= b; a`, obj.NewInteger(8))

	// Test bitwise shift right
	tt.add(`a = 8; b = 3; a >> b`, obj.NewInteger(1))
	tt.add(`a = 8; b = 3; a >>= b; a`, obj.NewInteger(1))

	// Test call
	tt.add(`add = fn(a, b) { a + b }; add(1, 2)`, obj.NewInteger(3))

	// Test loop
	tt.add(`for i = 0; i < 10; ++i {}; i`, obj.NewInteger(10))

	// Test list
	tt.add(`a = [1, 2, 3, 4, 5]; a[3]`, obj.NewInteger(4))
	tt.add(`a = [1, 2, 3, 4, 5]; a[2] = 6; a[2]`, obj.NewInteger(6))
	tt.add(`a = [1, 2, 3, 4, 5]; len(a)`, obj.NewInteger(5))
	tt.add(`a = [1, 2, 3, 4, 5]; a = append(a, 6); a[5]`, obj.NewInteger(6))

	// Test map
	tt.add(`a = {"key1": "value1", "key2": "value2"}; a["key1"]`, obj.NewString("value1"))
	tt.add(`a = {}; a["key1"] = "value1"; a["key1"]`, obj.NewString("value1"))
	tt.add(`a = {"key1": "value1"}; a["key1"] = "new_value1"; a["key1"]`, obj.NewString("new_value1"))

	// Test string interpolation
	tt.add(`a = 123; b = 456; "test {a} and {b}"`, obj.NewString("test 123 and 456"))

	tt.run(t)
}
