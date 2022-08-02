//go:build windows
// +build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

func evalREPL() {
	var (
		env    = obj.NewEnv()
		reader = bufio.NewReader(os.Stdin)
	)

	for {
		fmt.Print(">>> ")
		input, err := reader.ReadString('\n')
		check(err)

		input = strings.TrimRight(input, " \n")
		if len(input) > 0 && input[len(input)-1] == '{' {
			input, err = acceptUntil(reader, input, "\n\n")
			check(err)
		}

		res, errs := parser.Parse(input)
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			continue
		}

		if val := res.Eval(env); val != nil && val != obj.NullObj {
			fmt.Println(val)
		}
	}
}

func vmREPL() {
	var (
		consts      []obj.Object
		globals     = make([]obj.Object, vm.GlobalSize)
		symbolTable = compiler.NewSymbolTable()
		reader      = bufio.NewReader(os.Stdin)
	)

	for i, b := range obj.Builtins {
		symbolTable.DefineBuiltin(i, b.Name)
	}

	for {
		fmt.Print(">>> ")
		input, err := reader.ReadString('\n')
		check(err)

		input = strings.TrimRight(input, " \n")
		if len(input) > 0 && input[len(input)-1] == '{' {
			input, err = acceptUntil(reader, input, "\n\n")
			check(err)
		}

		res, errs := parser.Parse(input)
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			continue
		}

		c := compiler.NewWithState(symbolTable, &consts)
		if err := c.Compile(res); err != nil {
			fmt.Println(err)
			continue
		}
		tvm := vm.NewWithGlobalStore(c.Bytecode(), globals)

		if err := tvm.Run(); err != nil {
			fmt.Printf("runtime error: %v\n", err)
			continue
		}

		fmt.Println(tvm.LastPoppedStackElem())
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

func acceptUntil(r *bufio.Reader, start, end string) (string, error) {
	var buf strings.Builder

	buf.WriteString(start)
	buf.WriteRune('\n')
	for {
		fmt.Print("... ")
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}

		line = strings.TrimRight(line, " \n")
		buf.WriteString(line)
		buf.WriteRune('\n')

		if s := buf.String(); len(s) > len(end) && s[len(s)-len(end):] == end {
			break
		}
	}

	return buf.String(), nil
}
