// Command tau-rt is the runtime a bundled program is built on: it runs the
// program appended to it and nothing else. It carries the VM, the objects and
// the bytecode decoder, and not the lexer, the parser, the syntax tree or the
// compiler, which a program that is already compiled has no use for.
package main

import (
	"fmt"
	"os"

	"github.com/NicoNex/tau/internal/bundle"
	"github.com/NicoNex/tau/internal/vm"
)

func main() {
	if !bundle.HasEmbedded() {
		fmt.Fprintln(os.Stderr, "tau-rt: this is the runtime a bundled program is built on, it carries no program of its own")
		os.Exit(1)
	}

	vm.SetArgs(os.Args)
	if err := bundle.RunEmbedded(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
