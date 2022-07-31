package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/NicoNex/tau/compiler"
	"github.com/NicoNex/tau/obj"
	"github.com/NicoNex/tau/parser"
	"github.com/NicoNex/tau/vm"
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

func execFileVM(f string) {
	var bytecode *compiler.Bytecode

	if strings.HasSuffix(f, ".tauc") {
		file, err := os.Open(f)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		bytecode, err = decode(bufio.NewReader(file))
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		b := readFile(f)
		res, errs := parser.Parse(string(b))
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			return
		}

		c := compiler.New()
		if err := c.Compile(res); err != nil {
			fmt.Println(err)
			return
		}
		bytecode = c.Bytecode()
	}

	tvm := vm.New(bytecode)
	if err := tvm.Run(); err != nil {
		fmt.Printf("runtime error: %v\n", err)
		return
	}
}

func execFileEval(f string) {
	var env = obj.NewEnv()

	b := readFile(f)
	res, errs := parser.Parse(string(b))
	if len(errs) != 0 {
		for _, e := range errs {
			fmt.Println(e)
		}
		return
	}

	res.Eval(env)
}

func compileFiles(files []string) {
	for _, f := range files {
		b := readFile(f)

		res, errs := parser.Parse(string(b))
		if len(errs) != 0 {
			for _, e := range errs {
				fmt.Println(e)
			}
			return
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
}

func main() {
	var compile, useEval bool

	flag.BoolVar(&useEval, "eval", false, "Use the Tau eval function instead of the Tau VM. (slower)")
	flag.BoolVar(&compile, "c", false, "Compile a tau file into a '.tauc' bytecode file.")
	flag.Parse()

	switch {
	case compile:
		compileFiles(flag.Args())

	case flag.NArg() > 0:
		if useEval {
			execFileEval(flag.Arg(0))
		} else {
			execFileVM(flag.Arg(0))
		}

	default:
		if useEval {
			evalREPL()
		} else {
			vmREPL()
		}
	}
}
