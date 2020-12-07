// +build !windows

package main

import (
	"fmt"
	"io"
	"os"

	"tau/obj"
	"tau/parser"

	"golang.org/x/crypto/ssh/terminal"
)

func main() {
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
