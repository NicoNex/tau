package tau

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

func loadBuiltins(st *compiler.SymbolTable) *compiler.SymbolTable {
	for i, name := range obj.Builtins {
		st.DefineBuiltin(i, name)
	}
	return st
}

func SimpleVmREPL() {
	var (
		state   = vm.NewState()
		symbols = loadBuiltins(compiler.NewSymbolTable())
		reader  = bufio.NewReader(os.Stdin)
	)

	PrintVersionInfo(os.Stdout)
	for {
		fmt.Print(">>> ")
		input, err := reader.ReadString('\n')
		simpleCheck(err)

		input = strings.TrimRight(input, " \n")
		if len(input) > 0 && input[len(input)-1] == '{' {
			input, err = simpleAcceptUntil(reader, input, "\n\n")
			simpleCheck(err)
		}

		res, errs := parser.Parse("<stdin>", input)
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			continue
		}

		c := compiler.NewWithState(symbols, &vm.Consts)
		c.SetFileInfo("<stdin>", input)
		if err := c.Compile(res); err != nil {
			fmt.Println(err)
			continue
		}

		tvm := vm.NewWithState("<stdin>", c.Bytecode(), state)
		tvm.Run()
		state = tvm.State()
		tvm.Free()
	}
}

func simpleCheck(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}

func simpleAcceptUntil(r *bufio.Reader, start, end string) (string, error) {
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
