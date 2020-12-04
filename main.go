// +build !windows

package main

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh/terminal"
	"tau/parser"
)

func main() {
	initState, err := terminal.MakeRaw(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer terminal.Restore(0, initState)

	term := terminal.NewTerminal(os.Stdin, ">>> ")
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
			var tmp []interface{}

			for _, e := range errs {
				tmp = append(tmp, e)
			}
			fmt.Fprintln(term, tmp)
			continue
		}
		fmt.Fprintln(term, res.Eval())
	}
}
