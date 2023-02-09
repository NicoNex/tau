package main

import (
	"flag"
	"os"

	"github.com/NicoNex/tau"
)

func main() {
	var (
		compile bool
		useEval bool
		useFast bool
		version bool
		simple  bool
	)

	flag.BoolVar(&useFast, "f", false, "Use the fast implementation Tau VM (alpha).")
	flag.BoolVar(&useEval, "eval", false, "Use the Tau eval function instead of the Tau VM. (slower)")
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
		if useEval {
			tau.ExecFileEval(flag.Arg(0))
		} else if useFast {
			tau.ExecFileFastVM(flag.Arg(0))
		} else {
			tau.ExecFileVM(flag.Arg(0))
		}

	case simple:
		if useEval {
			tau.SimpleEvalREPL()
		} else {
			tau.SimpleVmREPL()
		}

	default:
		if useEval {
			tau.EvalREPL()
		} else {
			tau.VmREPL()
		}
	}
}
