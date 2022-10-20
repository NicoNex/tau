//go:build !windows
// +build !windows

package tau

import (
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

func EvalREPL() error {
	var env = obj.NewEnv("<stdin>")

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

		if val := res.Eval(env); val != nil && val != obj.NullObj {
			fmt.Fprintln(t, val)
		}
	}
}

func VmREPL() error {
	var state = vm.NewState()

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

		c := compiler.NewWithState(state.Symbols, &state.Consts)
		if err := c.Compile(res); err != nil {
			if ce, ok := err.(*compiler.CompilerError); ok {
				fmt.Fprintln(t, tauerr.New("<stdin>", input, ce.Pos(), ce.Error()))
			} else {
				fmt.Fprintln(t, err)
			}
			continue
		}

		tvm := vm.NewWithState("<stdin>", c.Bytecode(), state)
		if err := tvm.Run(); err != nil {
			fmt.Fprintf(t, "runtime error: %v\n", err)
			continue
		}

		if val := tvm.LastPoppedStackElem(); val != nil && val != obj.NullObj {
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
