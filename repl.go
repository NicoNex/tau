package tau

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

func VmREPL() error {
	var (
		state   = vm.NewState()
		symbols = loadBuiltins(compiler.NewSymbolTable())
	)

	initState, err := term.MakeRaw(0)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error opening terminal: %w", err)
	}
	defer term.Restore(0, initState)

	t := term.NewTerminal(os.Stdin, ">>> ")
	t.AutoCompleteCallback = autoComplete
	obj.Stdout = t

	PrintVersionInfo(t)
	for {
		input, err := t.ReadLine()
		check(t, initState, err)

		input = strings.TrimRight(input, " ")
		if len(input) > 0 && input[len(input)-1] == '{' {
			input, err = acceptUntil(t, input, "\n\n")
			check(t, initState, err)
		}

		res, errs := parser.Parse("<stdin>", input)
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Fprintln(t, e)
			}
			continue
		}

		c := compiler.NewWithState(symbols, &state.Consts)
		if err := c.Compile(res); err != nil {
			fmt.Fprintln(t, err)
			continue
		}

		tvm := vm.NewWithState("<stdin>", c.Bytecode(), state)
		if err := tvm.Run(); err != nil {
			fmt.Fprintf(t, "runtime error: %v\n", err)
			continue
		}

		if val := tvm.LastPoppedStackElem(); val != nil && obj.IsPrimitive(val) {
			fmt.Fprintln(t, val)
		}
	}
}

func autoComplete(line string, pos int, key rune) (newLine string, newPos int, ok bool) {
	if key == '\t' {
		return line + "    ", pos + 4, true
	}
	return
}

func check(t *term.Terminal, initState *term.State, err error) {
	if err != nil {
		// Quit without error on Ctrl^D.
		if err != io.EOF {
			fmt.Fprintln(t, err)
		}
		term.Restore(0, initState)
		fmt.Println()
		os.Exit(0)
	}
}

func acceptUntil(t *term.Terminal, start, end string) (string, error) {
	var buf strings.Builder

	buf.WriteString(start)
	buf.WriteRune('\n')
	t.SetPrompt("... ")
	defer t.SetPrompt(">>> ")

	for {
		line, err := t.ReadLine()
		if err != nil {
			return "", err
		}

		line = strings.TrimRight(line, " ")
		buf.WriteString(line)
		buf.WriteRune('\n')

		if s := buf.String(); len(s) > len(end) && s[len(s)-len(end):] == end {
			break
		}
	}

	return buf.String(), nil
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

		c := compiler.NewWithState(symbols, &state.Consts)
		c.SetFileInfo("<stdin>", input)
		if err := c.Compile(res); err != nil {
			fmt.Println(err)
			continue
		}
		tvm := vm.NewWithState("<stdin>", c.Bytecode(), state)

		if err := tvm.Run(); err != nil {
			fmt.Printf("runtime error: %v\n", err)
			continue
		}

		if val := tvm.LastPoppedStackElem(); val != nil && obj.IsPrimitive(val) {
			fmt.Println(val)
		}
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

func loadBuiltins(st *compiler.SymbolTable) *compiler.SymbolTable {
	for i, b := range obj.Builtins {
		st.DefineBuiltin(i, b.Name)
	}
	return st
}
