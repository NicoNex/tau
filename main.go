package tau

import (
	"flag"
	"fmt"
	"io"
)

const binaryName = "tau"

type Option struct {
	helpFlag    bool
	compileFlag bool
	evalFlag    bool
}

func NewFlagSet() (*flag.FlagSet, *Option) {
	ret := flag.NewFlagSet(binaryName, flag.ExitOnError)

	ret.Usage = func() {
		fmt.Printf("%s [OPTIONS] [FILE]", binaryName)
		ret.PrintDefaults()
	}

	var opt Option

	ret.BoolVar(&opt.helpFlag, "help", false, "show this message")
	flag.BoolVar(&opt.evalFlag, "eval", false, "Use the Tau eval function instead of the Tau VM. (slower)")
	flag.BoolVar(&opt.compileFlag, "c", false, "Compile a tau file into a '.tauc' bytecode file.")

	return ret, &opt
}

func Main(stdout io.Writer, args []string) error {
	flagSet, opt := NewFlagSet()
	flagSet.Parse(args)

	if flagSet.NArg() < 1 || opt.helpFlag {
		flagSet.Usage()
		return nil
	}

	return tau(stdout, flagSet.Args(), opt)
}

func tau(w io.Writer, args []string, opt *Option) error {
	if opt.compileFlag {
		return CompileFiles(args)
	}

	if len(args) > 0 {
		if opt.evalFlag {
			return ExecFileEval(args[0])
		}
		return ExecFileVM(args[0])
	}

	if opt.evalFlag {
		return EvalREPL()
	}
	return VmREPL()
}
