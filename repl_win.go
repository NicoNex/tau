//go:build windows
// +build windows

package tau

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/tauerr"
	"github.com/NicoNex/tau/internal/vm"
)

func EvalREPL() {
	var (
		env    = obj.NewEnv("<stdin>")
		reader = bufio.NewReader(os.Stdin)
	)

	PrintVersionInfo(os.Stdout)
	for {
		fmt.Print(">>> ")
		input, err := reader.ReadString('\n')
		check(err)

		input = strings.TrimRight(input, " \n")
		if len(input) > 0 && input[len(input)-1] == '{' {
			input, err = acceptUntil(reader, input, "\n\n")
			check(err)
		}

		res, errs := parser.Parse("<stdin>", input)
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			continue
		}

		if val := res.Eval(env); val != nil && val != obj.NullObj {
			fmt.Println(val)
		}
	}
}

func VmREPL() {
	var (
		state  = vm.NewState()
		reader = bufio.NewReader(os.Stdin)
	)

	PrintVersionInfo(os.Stdout)
	for {
		fmt.Print(">>> ")
		input, err := reader.ReadString('\n')
		check(err)

		input = strings.TrimRight(input, " \n")
		if len(input) > 0 && input[len(input)-1] == '{' {
			input, err = acceptUntil(reader, input, "\n\n")
			check(err)
		}

		res, errs := parser.Parse("<stdin>", input)
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			continue
		}

		c := compiler.NewWithState(state.Symbols, &state.Consts)
		c.SetFileContent(input)
		if err := c.Compile(res); err != nil {
			if ce, ok := err.(*compiler.CompilerError); ok {
				fmt.Println(tauerr.New("<stdin>", input, ce.Pos(), ce.Error()))
			} else {
				fmt.Println(err)
			}
			continue
		}
		tvm := vm.NewWithState("<stdin>", c.Bytecode(), state)

		if err := tvm.Run(); err != nil {
			fmt.Printf("runtime error: %v\n", err)
			continue
		}

		fmt.Println(tvm.LastPoppedStackElem())
	}
}

func check(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

func acceptUntil(r *bufio.Reader, start, end string) (string, error) {
	var buf strings.Builder

	buf.WriteString(start)
	buf.WriteRune('\n')
	for {
		fmt.Print("... ")
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}

		line = strings.TrimRight(line, " \n")
		buf.WriteString(line)
		buf.WriteRune('\n')

		if s := buf.String(); len(s) > len(end) && s[len(s)-len(end):] == end {
			break
		}
	}

	return buf.String(), nil
}
