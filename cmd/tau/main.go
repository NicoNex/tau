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

// bundle writes the program and its imports into a copy of the interpreter,
// which makes an executable that runs on its own.
func bundle() error {
	opt := parseBundleOpts()
	if opt.file == "" {
		usageBundle()
		return errUsage
	}
	return tau.BuildExecutable(opt.file, opt.output)
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
	case "bundle":
		usageBundle()
	case "test":
		usageTest()
	case "fmt":
		usageFmt()
	case "repl":
		usageRepl()
	case "get":
		usageGet()
	case "mod":
		usageMod()
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
	// An interpreter with a program appended to it is that program: the
	// arguments belong to it, not to the commands this binary would otherwise
	// answer to.
	if tau.HasEmbeddedProgram() {
		tau.SetArgs(os.Args)
		check(tau.RunEmbedded())
		return
	}

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
	case "bundle":
		check(bundle())
	case "test":
		check(test())
	case "fmt":
		check(format())
	case "repl":
		check(repl())
	case "get":
		check(get())
	case "mod":
		check(modcmd())
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

// get fetches a module and writes it into tau.mod.
func get() error {
	if len(os.Args) < 3 {
		usageGet()
		return nil
	}
	return tau.Get(os.Args[2])
}

// modcmd is the family that looks after the manifest itself.
func modcmd() error {
	if len(os.Args) < 3 {
		usageMod()
		return nil
	}

	switch sub := os.Args[2]; sub {
	case "init":
		var path string
		if len(os.Args) > 3 {
			path = os.Args[3]
		}
		return tau.ModInit(path)
	case "tidy":
		return tau.ModTidy()
	case "download":
		return tau.ModDownload()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", "mod "+sub)
		usageMod()
		os.Exit(1)
		return nil
	}
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
