package tau

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/NicoNex/tau/internal/ast"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

const TauVersion = "v1.5.0"

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
	return tauDecode(b), nil
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
		bytecode = tauDecode(readFile(f))
	} else {
		bytecode, err = compile(f)
	}

	if err != nil {
		fmt.Println(err)
		return
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
		writeFile(f[:len(f)-len(ext)]+".tauc", tauEncode(c.Bytecode()))
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
