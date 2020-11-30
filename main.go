// +build !windows

package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/NicoNex/calc/parser"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	OK int = iota
	ERR
)

var initState *terminal.State

func exit(code int, a ...interface{}) {
	terminal.Restore(0, initState)
	fmt.Println(a...)
	os.Exit(code)
}

func main() {
	if len(os.Args) > 1 {
		input := strings.Join(os.Args[1:], " ")
		if res, err := parser.Parse(input); err == nil {
			fmt.Println(res.Eval())
		} else {
			fmt.Println(err)
		}
		return
	}

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

		res, err := parser.Parse(input)
		if err != nil {
			fmt.Fprintln(term, err)
			continue
		}
		fmt.Fprintln(term, res.Eval())
	}
}
