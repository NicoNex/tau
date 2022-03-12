//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"io"
	"os"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
	"github.com/NicoNex/tau/parser"
	"github.com/NicoNex/tau/vm"
	"golang.org/x/term"
)

func evalREPL() {
	var env = obj.NewEnv()

	initState, err := term.MakeRaw(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer term.Restore(0, initState)

	t := term.NewTerminal(os.Stdin, ">>> ")
	obj.Stdout = t

	for {
		input, err := t.ReadLine()
		if err != nil {
			// Quit without error on Ctrl^D.
			if err != io.EOF {
				fmt.Fprintln(t, err)
			}
			return
		}

		res, errs := parser.Parse(input)
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Fprintln(t, e)
			}
			continue
		}

		if val := res.Eval(env); val != nil && val != obj.NullObj {
			fmt.Fprintln(t, val)
		}
	}
}

func vmREPL() {
	var (
		consts      []obj.Object
		globals     = make([]obj.Object, vm.GlobalSize)
		symbolTable = compiler.NewSymbolTable()
	)

	for i, b := range obj.Builtins {
		symbolTable.DefineBuiltin(i, b.Name)
	}

	initState, err := term.MakeRaw(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer term.Restore(0, initState)

	t := term.NewTerminal(os.Stdin, ">>> ")
	obj.Stdout = t

	for {
		input, err := t.ReadLine()
		if err != nil {
			// Quit without error on Ctrl^D.
			if err != io.EOF {
				fmt.Fprintln(t, err)
			}
			return
		}

		res, errs := parser.Parse(input)
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Fprintln(t, e)
			}
			continue
		}

		c := compiler.NewWithState(symbolTable, &consts)
		if err := c.Compile(res); err != nil {
			fmt.Fprintln(t, err)
			continue
		}
		tvm := vm.NewWithGlobalStore(c.Bytecode(), globals)

		if err := tvm.Run(); err != nil {
			fmt.Fprintf(t, "runtime error: %v\n", err)
			continue
		}

		fmt.Fprintln(t, tvm.LastPoppedStackElem())
	}
}
