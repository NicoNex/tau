package tau

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/NicoNex/tau/internal/ast"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

const TauVersion = "v2.0.15"

var ErrParseError = errors.New("error: parse error")

func readFile(fname string) []byte {
	b, err := os.ReadFile(fname)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return b
}

func writeFile(fname string, cont []byte) {
	if err := os.WriteFile(fname, cont, 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func precompiledBytecode(path string) (compiler.Bytecode, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return compiler.Bytecode{}, fmt.Errorf("error opening file %q: %w", path, err)
	}
	return compiler.DecodeBytecode(b), nil
}

func compile(path string) (bc compiler.Bytecode, err error) {
	input := string(readFile(path))
	res, errs := parser.Parse(path, input)
	if len(errs) > 0 {
		var buf strings.Builder

		for _, e := range errs {
			buf.WriteString(e.Error())
			buf.WriteByte('\n')
		}
		return compiler.Bytecode{}, errors.New(buf.String())
	}

	c := compiler.New()
	c.SetFileInfo(path, input)
	if err = c.Compile(res); err != nil {
		return
	}

	return c.Bytecode(), nil
}

func ExecFileVM(f string) (err error) {
	var bytecode compiler.Bytecode

	if filepath.Ext(f) == ".tauc" {
		raw := readFile(f)

		if IsBundle(raw) {
			var clean func()

			if bytecode, clean, err = openBundle(raw); err != nil {
				fmt.Println(err)
				return err
			}
			defer clean()
		} else {
			bytecode = compiler.DecodeBytecode(raw)
		}
	} else {
		if bytecode, err = compile(f); err != nil {
			fmt.Println(err)
			return err
		}
	}

	tvm := vm.New(f, bytecode)
	tvm.Run()
	return nil
}

// SetArgs hands the command line to the program about to run, which reads it
// back as os.Args. It travels in the environment rather than as a builtin so
// that the language keeps its small set of globals: the os module is the only
// one that knows these variables exist, and it clears them once read.
//
// ponytail: one variable per argument because an argument may hold anything
// but a NUL, so no single separator would be safe to join them with.
func SetArgs(args []string) {
	os.Setenv("TAU_ARGC", strconv.Itoa(len(args)))
	for i, a := range args {
		os.Setenv(fmt.Sprintf("TAU_ARG%d", i), a)
	}
}

// CompileFiles compiles each file into a self contained '.tauc' bundle. With
// out empty each bundle is written next to its source.
func CompileFiles(files []string, out string) error {
	if out != "" && len(files) > 1 {
		return errors.New("build: -o takes a single input file")
	}

	for _, f := range files {
		bundle, err := Bundle(f)
		if err != nil {
			return err
		}

		dst := out
		if dst == "" {
			ext := filepath.Ext(f)
			dst = f[:len(f)-len(ext)] + ".tauc"
		}
		writeFile(dst, bundle)
	}

	return nil
}

func PrintVersionInfo(w io.Writer) {
	fmt.Fprintf(w, "Tau %s on %s\n", TauVersion, strings.Title(runtime.GOOS))
}

func Parse(src string) (ast.Node, error) {
	tree, errs := parser.Parse("<input>", src)
	if len(errs) > 0 {
		var buf strings.Builder

		buf.WriteString("parser error:\n")
		for _, e := range errs {
			buf.WriteString(e.Error())
			buf.WriteByte('\n')
		}

		return nil, errors.New(buf.String())
	}

	return tree, nil
}

// TestFiles runs every *_test.tau in the given paths, a directory standing
// for the test files it holds, and reports how many of them failed. Each file
// runs in its own process the way `go test` does with its packages, so a test
// that exits or crashes takes down only itself.
func TestFiles(paths []string) error {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, p)
			continue
		}

		matches, err := filepath.Glob(filepath.Join(p, "*_test.tau"))
		if err != nil {
			return err
		}
		files = append(files, matches...)
	}

	if len(files) == 0 {
		return errors.New("no test files found")
	}
	sort.Strings(files)

	self, err := os.Executable()
	if err != nil {
		self = os.Args[0]
	}

	var failed int
	for _, f := range files {
		fmt.Printf("=== %s\n", f)

		cmd := exec.Command(self, f)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			failed++
		}
	}

	fmt.Println()
	if failed > 0 {
		return fmt.Errorf("%d of %d test files failed", failed, len(files))
	}
	fmt.Printf("ok      %d test files passed\n", len(files))
	return nil
}

// FormatFiles rewrites the given tau files in the canonical style, walking
// directories for '.tau' files. With write the files are rewritten in place,
// with list only the names of the ones that differ are printed, otherwise the
// formatted source goes to standard output.
func FormatFiles(paths []string, write, list bool) error {
	return errors.New("fmt: not implemented yet")
}
