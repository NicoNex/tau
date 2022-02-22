//go:build windows
// +build windows

package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
	"github.com/NicoNex/tau/parser"
	"github.com/NicoNex/tau/vm"
)

func evalREPL() {
	var (
		env    = obj.NewEnv()
		reader = bufio.NewReader(os.Stdin)
	)

	for {
		fmt.Print(">>> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
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
		if err != nil {
			fmt.Println(err)
			return
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
