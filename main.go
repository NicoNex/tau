// +build !windows

package main

import (
	"fmt"
	// "io"
	// "os"
	// "strings"

	// "tau/obj"
	"tau/ast"
	// "tau/parser"
	// "golang.org/x/crypto/ssh/terminal"
)

func main() {
	l := ast.NewInteger(3)
	r := ast.NewInteger(5)
	sum := ast.NewPlus(l, r)

	fmt.Println(sum.Eval())


	// initState, err := terminal.MakeRaw(0)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer terminal.Restore(0, initState)

	// term := terminal.NewTerminal(os.Stdin, ">>> ")
	// for {
	// 	input, err := term.ReadLine()
	// 	if err != nil {
	// 		// Quit without error on Ctrl^D.
	// 		if err != io.EOF {
	// 			fmt.Println(err)
	// 		}
	// 		return
	// 	}

	// 	res, err := parser.Parse(input)
	// 	if err != nil {
	// 		fmt.Fprintln(term, err)
	// 		continue
	// 	}
	// 	fmt.Fprintln(term, res.Eval())
	// }
}
