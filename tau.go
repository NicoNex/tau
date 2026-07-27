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
		bytecode = compiler.DecodeBytecode(readFile(f))
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

func CompileFiles(files []string) error {
	for _, f := range files {
		b := readFile(f)

		res, errs := parser.Parse(f, string(b))
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			return ErrParseError
		}

		c := compiler.New()
		c.SetFileInfo(f, string(b))
		if err := c.Compile(res); err != nil {
			fmt.Println(err)
			continue
		}
		ext := filepath.Ext(f)
		writeFile(f[:len(f)-len(ext)]+".tauc", c.Bytecode().Encode())
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
