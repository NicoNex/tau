//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/NicoNex/tau/obj"
	"github.com/NicoNex/tau/parser"
	"golang.org/x/term"
)

func repl() {
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
		if val := res.Eval(env); val != nil && val != obj.NullObj {
			fmt.Fprintln(t, val)
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
