//go:build windows

package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/NicoNex/tau/obj"
	"github.com/NicoNex/tau/parser"
)

func repl() {
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

		val := res.Eval(env)
		if val != obj.NullObj && val != nil {
			fmt.Println(val)
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
