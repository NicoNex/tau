package tau

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
	"golang.org/x/term"
)

// TODO: somehow redirect stdout to the terminal.
func REPL() error {
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
	redirectStdout(t)

	PrintVersionInfo(t)
	for {
		input, err := t.ReadLine()
		check(t, initState, err)

		input = strings.TrimRight(input, " ")
		if input == "" {
			continue
		} else if len(input) > 0 && input[len(input)-1] == '{' {
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

		c := compiler.NewWithState(symbols, &vm.Consts)
		c.SetFileInfo("<stdin>", input)
		if err := c.Compile(res); err != nil {
			fmt.Fprintln(t, err)
			continue
		}

		tvm := vm.NewWithState("<stdin>", c.Bytecode(), state)
		tvm.Run()
		state = tvm.State()
		tvm.Free()
	}
}

func redirectStdout(w io.Writer) {
	pr, pw, err := os.Pipe()
	if err != nil {
		fmt.Println("Error creating pipe:", err)
		return
	}
	// Set the pipe writer as the stdout
	syscall.Dup2(int(pw.Fd()), syscall.Stdout)

	go func() {
		var buf = make([]byte, 4096)

		for {
			n, err := pr.Read(buf)
			if err != nil {
				if err != io.EOF {
					fmt.Println("error reading from pipe:", err)
				}
				break
			}

			// Write the captured output to the provided writer
			_, err = w.Write(buf[:n])
			if err != nil {
				fmt.Println("error writing to writer:", err)
				break
			}
		}
		pr.Close()
	}()
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

func SimpleREPL() {
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

func loadBuiltins(st *compiler.SymbolTable) *compiler.SymbolTable {
	for i, name := range obj.Builtins {
		st.DefineBuiltin(i, name)
	}
	return st
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
