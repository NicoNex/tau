package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/NicoNex/tau"
)

// run executes a tau file, handing it the arguments that follow so the
// program can read them through os.Args.
func run() error {
	opt := parseRunOpts()
	if opt.path == "" {
		usageRun()
		return errUsage
	}

	tau.SetArgs(append([]string{opt.path}, opt.args...))
	return tau.ExecFileVM(opt.path)
}

// build compiles tau files down to .tauc bytecode.
func build() error {
	opt := parseBuildOpts()
	if len(opt.files) == 0 {
		usageBuild()
		return errUsage
	}
	return tau.CompileFiles(opt.files, opt.output)
}

// test runs the *_test.tau files, like `go test` does.
func test() error {
	return tau.TestFiles(parseTestOpts().paths)
}

// format rewrites tau files in the canonical style, like `go fmt` does.
func format() error {
	opt := parseFmtOpts()
	if len(opt.paths) == 0 {
		opt.paths = []string{"."}
	}
	return tau.FormatFiles(opt.paths, opt.write, opt.list)
}

func version() error {
	tau.PrintVersionInfo(os.Stdout)
	return nil
}

func help() error {
	if len(os.Args) < 3 {
		usageGeneral()
		return nil
	}

	switch cmd := os.Args[2]; cmd {
	case "run":
		usageRun()
	case "build":
		usageBuild()
	case "test":
		usageTest()
	case "fmt":
		usageFmt()
	case "repl":
		usageRepl()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usageGeneral()
	}
	return nil
}

// repl opens the interactive prompt, falling back to the simple one where a
// terminal isn't available.
func repl() error {
	opt := parseReplOpts()
	if opt.simple || runtime.GOOS == "windows" {
		tau.SimpleREPL()
	} else {
		tau.REPL()
	}
	return nil
}

func check(err error) {
	if err == nil {
		return
	}
	if err != errUsage {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		tau.SetArgs(os.Args)
		tau.REPL()
		return
	}

	switch cmd := os.Args[1]; cmd {
	case "run":
		check(run())
	case "build":
		check(build())
	case "test":
		check(test())
	case "fmt":
		check(format())
	case "repl":
		check(repl())
	case "version", "-v", "--version":
		check(version())
	case "help", "-h", "--help":
		check(help())
	default:
		// "tau file.tau args..." stays a shorthand for "tau run", so a
		// shebang line and every script written so far keeps working.
		if isFile(cmd) {
			tau.SetArgs(os.Args[1:])
			check(tau.ExecFileVM(cmd))
			return
		}
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", cmd)
		usageGeneral()
		os.Exit(1)
	}
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
