package tau

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

var (
	ErrParseError = errors.New("error: parse error")
)

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
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return b
}

func writeFile(fname string, cont []byte) {
	if err := ioutil.WriteFile(fname, cont, 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func ExecFileVM(f string) error {
	var bytecode *compiler.Bytecode

	if strings.HasSuffix(f, ".tauc") {
		file, err := os.Open(f)
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("error opening file %q: %w", f, err)
		}
		defer file.Close()

		bytecode, err = decode(bufio.NewReader(file))
		if err != nil {
			fmt.Println(err)
			return fmt.Errorf("error decoding bytecode: %w", err)
		}
	} else {
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
			return fmt.Errorf("error during compilation: %w", err)
		}
		bytecode = c.Bytecode()
	}

	tvm := vm.New(bytecode)
	if err := tvm.Run(); err != nil {
		fmt.Printf("runtime error: %v\n", err)
		return fmt.Errorf("runtime error: %w", err)
	}

	return nil
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
