//go:build windows
// +build windows

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
	"github.com/NicoNex/tau/parser"
	"github.com/NicoNex/tau/vm"
)

var useVM bool

func repl() {
	var (
		env         *obj.Env
		consts      []obj.Object
		globals     []obj.Object
		symbolTable *compiler.SymbolTable
		reader      = bufio.NewReader(os.Stdin)
	)

	if useVM {
		globals = make([]obj.Object, vm.GlobalSize)
		symbolTable = compiler.NewSymbolTable()
		for i, b := range obj.Builtins {
			symbolTable.DefineBuiltin(i, b.Name)
		}
	} else {
		env = obj.NewEnv()
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

		if useVM {
			c := compiler.NewWithState(symbolTable, consts)
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
			continue
		}

		val := res.Eval(env)
		if val != obj.NullObj && val != nil {
			fmt.Println(val)
		}
	}
}

func main() {
	flag.BoolVar(&useVM, "vm", false, "Use the Tau VM instead of eval method.")
	flag.Parse()

	if flag.NArg() > 0 {
		var env = obj.NewEnv()

		b, err := ioutil.ReadFile(flag.Arg(0))
		if err != nil {
			fmt.Println(err)
			return
		}

		res, errs := parser.Parse(string(b))
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			return
		}

		val := res.Eval(env)
		if val != obj.NullObj && val != nil {
			fmt.Println(val)
		}
	} else {
		repl()
	}
}
