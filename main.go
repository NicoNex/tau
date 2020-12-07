// +build !windows

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/NicoNex/tau/obj"
	"github.com/NicoNex/tau/parser"

	"golang.org/x/crypto/ssh/terminal"
)

func repl() {
	var env = obj.NewEnv()

	initState, err := terminal.MakeRaw(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer terminal.Restore(0, initState)

	term := terminal.NewTerminal(os.Stdin, ">>> ")
	obj.Stdout = term
	for {
		input, err := term.ReadLine()
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
				fmt.Fprintln(term, e)
			}
			continue
		}

		val := res.Eval(env)
		if val != obj.NullObj && val != nil {
			fmt.Fprintln(term, val)
		}
	}
}

func main() {
	if len(os.Args) > 1 {
		var env = obj.NewEnv()

		b, err := ioutil.ReadFile(os.Args[1])
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
