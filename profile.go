//go:build ignore

package main

import (
	"flag"
	"os"
	"runtime/pprof"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"

	_ "github.com/ianlancetaylor/cgosymbolizer"
)

const fib = `
fib = fn(n) {
	if n < 2 {
		return n
	}
	fib(n-1) + fib(n-2)
}

println(fib(40))`

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func code(path string) string {
	b, err := os.ReadFile(path)
	check(err)

	return string(b)
}

func fileOrDefault() string {
	if flag.NArg() > 0 {
		return code(flag.Arg(0))
	}
	return fib
}

func main() {
	cpuf, err := os.Create("cpu.prof")
	check(err)

	tree, errs := parser.Parse("<profiler>", fileOrDefault())
	if len(errs) > 0 {
		panic("parser errors")
	}

	c := compiler.New()
	c.SetFileInfo("<profiler>", fib)
	check(c.Compile(tree))

	check(pprof.StartCPUProfile(cpuf))
	defer pprof.StopCPUProfile()
	tvm := vm.New("<profiler>", c.Bytecode())
	tvm.Run()
}
