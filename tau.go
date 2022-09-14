package tau

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

const (
	TauVersion = "v1.2.9"
	GoVersion  = "go version go1.19 linux/amd64"
)

var ErrParseError = errors.New("error: parse error")

func encode(bcode *compiler.Bytecode) ([]byte, error) {
	var (
		buf bytes.Buffer
		enc = gob.NewEncoder(&buf)
	)

	for _, c := range bcode.Constants {
		gob.Register(c)
	}

	if err := enc.Encode(bcode); err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}

func decode(r io.Reader) (*compiler.Bytecode, error) {
	var (
		b   *compiler.Bytecode
		dec = gob.NewDecoder(r)
	)

	gob.Register(obj.NewInteger(0))
	gob.Register(obj.NewBoolean(false))
	gob.Register(obj.NewNull())
	gob.Register(obj.NewClass())
	gob.Register(obj.NewReturn(nil))
	gob.Register(obj.NewFloat(0))
	gob.Register(obj.NewList())
	gob.Register(obj.NewMap())
	gob.Register(obj.NewString(""))
	gob.Register(obj.NewError(""))
	gob.Register(obj.NewClosure(nil, []obj.Object{}))
	gob.Register(obj.Builtin(func(arg ...obj.Object) obj.Object { return nil }))
	gob.Register(obj.NewFunction([]string{}, obj.NewEnv(), nil))

	return b, dec.Decode(&b)
}

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

func precompiledBytecode(path string) (*compiler.Bytecode, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("error opening file %q: %w", path, err)
	}
	defer file.Close()
	return decode(bufio.NewReader(file))
}

func compile(path string) (*compiler.Bytecode, error) {
	b := readFile(path)
	res, errs := parser.Parse(string(b))
	if len(errs) > 0 {
		var buf strings.Builder

		buf.WriteString("error parsing module ")
		buf.WriteString(path)
		buf.WriteRune(':')

		for _, e := range errs {
			buf.WriteRune('\t')
			buf.WriteString(e)
		}

		return nil, errors.New(buf.String())
	}

	c := compiler.New()
	if err := c.Compile(res); err != nil {
		return nil, err
	}

	return c.Bytecode(), nil
}

func ExecFileVM(f string) (err error) {
	var bytecode *compiler.Bytecode

	if filepath.Ext(f) == ".tauc" {
		bytecode, err = precompiledBytecode(f)
	} else {
		bytecode, err = compile(f)
	}

	if err != nil {
		fmt.Println(err)
		return
	}

	tvm := vm.New(bytecode)
	if err = tvm.Run(); err != nil {
		fmt.Println(err)
		return
	}

	return
}

func ExecFileEval(f string) error {
	var env = obj.NewEnv()

	b := readFile(f)
	res, errs := parser.Parse(string(b))
	if len(errs) != 0 {
		for _, e := range errs {
			fmt.Println(e)
		}
		return ErrParseError
	}

	res.Eval(env)

	return nil
}

func CompileFiles(files []string) error {
	for _, f := range files {
		b := readFile(f)

		res, errs := parser.Parse(string(b))
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			return ErrParseError
		}

		c := compiler.New()
		if err := c.Compile(res); err != nil {
			fmt.Println(err)
			continue
		}

		cnt, err := encode(c.Bytecode())
		if err != nil {
			fmt.Println(err)
			continue
		}

		ext := filepath.Ext(f)
		writeFile(f[:len(f)-len(ext)]+".tauc", cnt)
	}

	return nil
}

func PrintVersionInfo(w io.Writer) {
	fmt.Fprintf(w, "Tau %s [%s] on %s\n", TauVersion, GoVersion, runtime.GOOS)
}
