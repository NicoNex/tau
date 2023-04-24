package main

import (
	"flag"
	"os"

	"github.com/NicoNex/tau"
)

func main() {
	var (
		compile bool
		version bool
		simple  bool
	)

	flag.BoolVar(&compile, "c", false, "Compile a tau file into a '.tauc' bytecode file.")
	flag.BoolVar(&simple, "s", false, "Use simple REPL instead of opening a terminal.")
	flag.BoolVar(&version, "v", false, "Print Tau version information.")
	flag.Parse()

	switch {
	case compile:
		tau.CompileFiles(flag.Args())
	case version:
		tau.PrintVersionInfo(os.Stdout)
	case flag.NArg() > 0:
		tau.ExecFileVM(flag.Arg(0))
	case simple:
		tau.SimpleVmREPL()
	default:
		tau.VmREPL()
	}
}
