package tau

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/NicoNex/tau/internal/ast"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/vm"
)

const TauVersion = "v1.5.0"

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
	gob.Register(obj.NewFunction([]string{}, obj.NewEnv(""), nil))

	return b, dec.Decode(&b)
}

func encodeObjects(buf bytes.Buffer, objs []obj.Object) (err error) {
loop:
	for _, o := range objs {
		if err != nil {
			return
		}

		buf.WriteByte(byte(o.Type()))

		switch o.Type() {
		case obj.NullType:
			continue loop
		case obj.BoolType:
			if o == obj.True {
				buf.WriteByte(1)
			} else {
				buf.WriteByte(0)
			}
		case obj.IntType:
			i := uint64(o.(obj.Integer))
			val := make([]byte, 8)
			binary.BigEndian.PutUint64(val, i)
			buf.Write(val)
		case obj.StringType:
			s := []byte(o.(obj.String))
			length := make([]byte, 4)
			binary.BigEndian.PutUint32(length, uint32(len(s)))
			buf.Write(length)
			buf.Write(s)
		case obj.ErrorType:
			e := []byte(o.(obj.Error))
			length := make([]byte, 4)
			binary.BigEndian.PutUint32(length, uint32(len(e)))
			buf.Write(length)
			buf.Write(e)
		case obj.FloatType:
			f := float64(o.(obj.Float))
			val := make([]byte, 8)
			binary.BigEndian.PutUint64(val, math.Float64bits(f))
			buf.Write(val)
		case obj.ClosureType:
			c := o.(*obj.Closure)
			length := make([]byte, 4)
			binary.BigEndian.PutUint32(length, uint32(len(c.Free)))
			buf.Write(length)
			err = encodeObjects(buf, c.Free)
			o = c.Fn
			fallthrough
		case obj.FunctionType:
			f := o.(*obj.CompiledFunction)
			data := make([]byte, 12)
			binary.BigEndian.PutUint32(data, uint32(f.NumParams))
			binary.BigEndian.PutUint32(data[4:], uint32(f.NumLocals))
			binary.BigEndian.PutUint32(data[8:], uint32(len(f.Instructions)))
			buf.Write(data)
			buf.Write(f.Instructions)
		case obj.ListType:
			l := o.(obj.List).Val()
			length := make([]byte, 4)
			binary.BigEndian.PutUint32(length, uint32(len(l)))
			buf.Write(length)
			err = encodeObjects(buf, l)
		default:
			return fmt.Errorf("unsupported encoding for type %v", o.Type())
		}
	}

	return
}

func tauEncode(bcode *compiler.Bytecode) ([]byte, error) {
	var buf bytes.Buffer

	// Write the length of the instructions on the first 4 bytes of the buffer.
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(bcode.Instructions)))
	buf.Write(length)

	// Write the encoded constants on the tail of the buffer.
	if err := encodeObjects(buf, bcode.Constants); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
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

func compile(path string) (bc *compiler.Bytecode, err error) {
	input := string(readFile(path))
	res, errs := parser.Parse(path, input)
	if len(errs) > 0 {
		var buf strings.Builder

		for _, e := range errs {
			buf.WriteString(e.Error())
			buf.WriteByte('\n')
		}
		return nil, errors.New(buf.String())
	}

	c := compiler.New()
	c.SetFileInfo(path, input)
	if err = c.Compile(res); err != nil {
		return
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

	tvm := vm.New(f, bytecode)
	if err = tvm.Run(); err != nil {
		fmt.Println(err)
		return
	}

	return
}

func ExecFileEval(f string) error {
	var env = obj.NewEnv(f)

	b := readFile(f)
	res, errs := parser.Parse(f, string(b))
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

		res, errs := parser.Parse(f, string(b))
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
