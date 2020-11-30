// +build windows

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/NicoNex/calc/parser"
)

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

	var reader = bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">>> ")
		string, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		ast, err := parser.Parse(string)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(ast.Eval())
	}
}
