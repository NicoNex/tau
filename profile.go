//go:build ignore

package main

import (
	"os"
	"runtime/pprof"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

const fib = `
fib = fn(n) {
	if n < 2 {
		return n
	}
	fib(n-1) + fib(n-2)
}

fib(35)`

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
	if len(os.Args) > 1 {
		return code(os.Args[1])
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
	check(c.Compile(tree))

	check(pprof.StartCPUProfile(cpuf))
	defer pprof.StopCPUProfile()

	tvm := vm.New("<profiler>", c.Bytecode())
	check(tvm.Run())
}
