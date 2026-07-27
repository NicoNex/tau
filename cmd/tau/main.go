package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

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

	// "tau test [path...]" runs the *_test.tau files, like `go test` does.
	if flag.Arg(0) == "test" {
		if err := tau.TestFiles(flag.Args()[1:]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	switch {
	case compile:
		tau.CompileFiles(flag.Args())
	case version:
		tau.PrintVersionInfo(os.Stdout)
	case flag.NArg() > 0:
		// A file that doesn't compile has to fail: a test runner reporting
		// success on a broken file is worse than useless.
		if err := tau.ExecFileVM(flag.Arg(0)); err != nil {
			os.Exit(1)
		}
	case simple || runtime.GOOS == "windows":
		tau.SimpleREPL()
	default:
		tau.REPL()
	}
}
