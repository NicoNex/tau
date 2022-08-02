package main

import (
	"flag"

	"github.com/NicoNex/tau"
)

func main() {
	var compile, useEval bool

	flag.BoolVar(&useEval, "eval", false, "Use the Tau eval function instead of the Tau VM. (slower)")
	flag.BoolVar(&compile, "c", false, "Compile a tau file into a '.tauc' bytecode file.")
	flag.Parse()

	switch {
	case compile:
		tau.CompileFiles(flag.Args())

	case flag.NArg() > 0:
		if useEval {
			tau.ExecFileEval(flag.Arg(0))
		} else {
			tau.ExecFileVM(flag.Arg(0))
		}

	default:
		if useEval {
			evalREPL()
		} else {
			vmREPL()
		}
	}
}
