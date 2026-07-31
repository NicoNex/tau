package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

// errUsage says the command was misused and its usage has already been
// printed, so main only has to exit non zero.
var errUsage = errors.New("usage")

type runOpt struct {
	path string
	args []string
}

type buildOpt struct {
	files  []string
	output string
}

type bundleOpt struct {
	file   string
	output string
}

type testOpt struct {
	paths []string
}

type fmtOpt struct {
	paths []string
	write bool
	list  bool
}

type replOpt struct {
	simple bool
}

func parseRunOpts() (opt runOpt) {
	cmd := flag.NewFlagSet("run", flag.ExitOnError)
	cmd.Usage = usageRun
	cmd.Parse(os.Args[2:])

	opt.path = cmd.Arg(0)
	if cmd.NArg() > 1 {
		opt.args = cmd.Args()[1:]
	}
	return
}

func parseBundleOpts() (opt bundleOpt) {
	cmd := flag.NewFlagSet("bundle", flag.ExitOnError)
	cmd.StringVar(&opt.output, "o", "", "Write the executable to the given file")
	cmd.StringVar(&opt.output, "out", "", "Write the executable to the given file (same as -o)")
	cmd.Usage = usageBundle
	cmd.Parse(os.Args[2:])

	opt.file = cmd.Arg(0)
	return
}

func parseBuildOpts() (opt buildOpt) {
	cmd := flag.NewFlagSet("build", flag.ExitOnError)
	cmd.StringVar(&opt.output, "o", "", "Write the bytecode to the given file")
	cmd.StringVar(&opt.output, "out", "", "Write the bytecode to the given file (same as -o)")
	cmd.Usage = usageBuild
	cmd.Parse(os.Args[2:])

	opt.files = cmd.Args()
	return
}

func parseTestOpts() (opt testOpt) {
	cmd := flag.NewFlagSet("test", flag.ExitOnError)
	cmd.Usage = usageTest
	cmd.Parse(os.Args[2:])

	opt.paths = cmd.Args()
	return
}

func parseFmtOpts() (opt fmtOpt) {
	cmd := flag.NewFlagSet("fmt", flag.ExitOnError)
	cmd.BoolVar(&opt.write, "w", false, "Write the result back to the source file")
	cmd.BoolVar(&opt.list, "l", false, "List the files whose formatting differs")
	cmd.Usage = usageFmt
	cmd.Parse(os.Args[2:])

	opt.paths = cmd.Args()
	return
}

func parseReplOpts() (opt replOpt) {
	cmd := flag.NewFlagSet("repl", flag.ExitOnError)
	cmd.BoolVar(&opt.simple, "s", false, "Use the simple REPL instead of opening a terminal")
	cmd.Usage = usageRepl
	cmd.Parse(os.Args[2:])

	return
}

func usageGeneral() {
	fmt.Fprintf(os.Stderr, `Usage: %s COMMAND [OPTIONS] ARGS

Tau is a dynamically typed, interpreted programming language.

Commands:
  run       Run a tau file
  build     Compile tau files into '.tauc' bytecode
  bundle    Compile a tau file into a standalone executable
  test      Run the tests of the given files or directories
  fmt       Format tau source files
  repl      Start the interactive prompt
  get       Fetch a module and add it to tau.mod
  mod       Look after tau.mod
  version   Print version information
  help      Display help for a command

Running '%s FILE [ARGS]' is a shorthand for '%s run FILE [ARGS]',
and running '%s' with no arguments starts the REPL.

Use '%s help COMMAND' for more information on a command.
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func usageRun() {
	fmt.Fprintf(os.Stderr, `Usage: %s run FILE [ARGS]

Run a tau file, either source or compiled bytecode. Anything after FILE is
passed to the program itself and read with os.Args.

Arguments:
  FILE      Path to a '.tau' source file or a '.tauc' bytecode file
  ARGS      Arguments for the program

Examples:
  %s run hello.tau
  %s run server.tau -port 8080
  %s hello.tau
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func usageBundle() {
	fmt.Fprintf(os.Stderr, `Usage: %s bundle [OPTIONS] FILE

Compile FILE and everything it imports into a single executable, a copy of
the interpreter with the program appended to it. The result runs where tau
is not installed and needs nothing else.

Without -o the executable takes the name of the source file without its
extension.

Options:
  -o, --out FILE    Write the executable to FILE

Arguments:
  FILE              Path to the '.tau' file to bundle

Examples:
  %s bundle main.tau
  %s bundle -o myapp main.tau
`, os.Args[0], os.Args[0], os.Args[0])
}

func usageBuild() {
	fmt.Fprintf(os.Stderr, `Usage: %s build [OPTIONS] FILE...

Compile tau source files into '.tauc' bytecode. Without -o each file is
written next to its source with the extension replaced.

Options:
  -o, --out FILE    Write the bytecode to FILE (only with a single input)

Arguments:
  FILE...           Paths to the '.tau' files to compile

Examples:
  %s build main.tau
  %s build -o app.tauc main.tau
  %s build *.tau
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func usageTest() {
	fmt.Fprintf(os.Stderr, `Usage: %s test [PATH...]

Run the '*_test.tau' files found in the given paths, a directory standing for
the test files it holds. With no path the current directory is used. Each file
runs in its own process, so a crashing test takes down only itself.

Arguments:
  PATH...   Files or directories to test (default: the current directory)

Examples:
  %s test
  %s test stdlib
  %s test stdlib/strings_test.tau
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func usageFmt() {
	fmt.Fprintf(os.Stderr, `Usage: %s fmt [OPTIONS] [PATH...]

Format tau source files in the canonical style: tabs for indentation, one
space around binary operators, comments kept where they are. A directory is
walked recursively for '.tau' files. With no path the current directory is
used.

Without options the formatted source is printed to standard output.

Options:
  -w        Write the result back to the source file
  -l        List the files whose formatting differs, without rewriting them

Arguments:
  PATH...   Files or directories to format (default: the current directory)

Examples:
  %s fmt main.tau
  %s fmt -w main.tau
  %s fmt -l stdlib
  %s fmt -w .
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func usageGet() {
	fmt.Fprintf(os.Stderr, `Usage: %s get PATH[@VERSION]

Fetch a module, add it to tau.mod and write its hash into tau.sum. Without a
version the highest one tagged vX.Y.Z is taken.

From v2 onwards the major is part of the path, so github.com/you/thing/v2 and
github.com/you/thing are two modules and a program may hold both.

Modules land in $TAUHOME, or ~/.tau when that is unset, under pkg/ and one
directory per version. Nothing is ever fetched while a program runs.

Fetching goes through git, which has to be installed. Building from what is
already there does not need it.

Examples:
  %s get github.com/NicoNex/example
  %s get github.com/NicoNex/example@v1.2.0
  %s get github.com/NicoNex/example/v2
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}

func usageMod() {
	fmt.Fprintf(os.Stderr, `Usage: %s mod COMMAND

Commands:
  init PATH   Write a tau.mod for the module in this directory
  tidy        Require what the source imports, drop what it does not
  download    Fetch everything tau.mod requires

Examples:
  %s mod init github.com/you/thing
  %s mod tidy
`, os.Args[0], os.Args[0], os.Args[0])
}

func usageRepl() {
	fmt.Fprintf(os.Stderr, `Usage: %s repl [OPTIONS]

Start the interactive prompt. Running '%s' with no arguments does the same.

Options:
  -s        Use the simple REPL instead of opening a terminal

Examples:
  %s repl
  %s repl -s
`, os.Args[0], os.Args[0], os.Args[0], os.Args[0])
}
