// +build windows

package main

import (
	"bufio"
	"fmt"
	"os"

	"tau/obj"
	"tau/parser"
)

func main() {
	var env = obj.NewEnv()
	var reader = bufio.NewReader(os.Stdin)

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
