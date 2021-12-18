//go:build !windows
// +build !windows

package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
	"github.com/NicoNex/tau/parser"
	"github.com/NicoNex/tau/vm"
	"golang.org/x/term"
)

var useVM bool

func repl() {
	var (
		env         *obj.Env
		consts      []obj.Object
		globals     []obj.Object
		symbolTable *compiler.SymbolTable
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
				fmt.Println(err)
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

		if useVM {
			c := compiler.NewWithState(symbolTable, consts)
			c.Compile(res)
			tvm := vm.NewWithGlobalStore(c.Bytecode(), globals)

			if err := tvm.Run(); err != nil {
				fmt.Fprintf(t, "runtime error: %v\n", err)
				continue
			}

			fmt.Fprintln(t, tvm.LastPoppedStackElem())
			continue
		}

		if val := res.Eval(env); val != nil && val != obj.NullObj {
			fmt.Fprintln(t, val)
		}
	}
}

func main() {
	flag.BoolVar(&useVM, "vm", false, "Use the Tau VM instead of eval method.")
	flag.Parse()

	if flag.NArg() > 0 {
		var (
			env         *obj.Env
			consts      []obj.Object
			globals     []obj.Object
			symbolTable *compiler.SymbolTable
		)

		if useVM {
			globals = make([]obj.Object, vm.GlobalSize)
			symbolTable = compiler.NewSymbolTable()
		} else {
			env = obj.NewEnv()
		}

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

		if useVM {
			c := compiler.NewWithState(symbolTable, consts)
			c.Compile(res)
			tvm := vm.NewWithGlobalStore(c.Bytecode(), globals)

			if err := tvm.Run(); err != nil {
				fmt.Printf("runtime error: %v\n", err)
				return
			}

			fmt.Println(tvm.LastPoppedStackElem())
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
