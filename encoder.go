package tau

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

var bin = binary.BigEndian

func encodeObjects(buf *bytes.Buffer, objs []obj.Object) (err error) {
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
			bin.PutUint64(val, i)
			buf.Write(val)
		case obj.StringType:
			s := []byte(o.(obj.String))
			length := make([]byte, 4)
			bin.PutUint32(length, uint32(len(s)))
			buf.Write(length)
			buf.Write(s)
		case obj.ErrorType:
			e := []byte(o.(obj.Error))
			length := make([]byte, 4)
			bin.PutUint32(length, uint32(len(e)))
			buf.Write(length)
			buf.Write(e)
		case obj.FloatType:
			f := float64(o.(obj.Float))
			val := make([]byte, 8)
			bin.PutUint64(val, math.Float64bits(f))
			buf.Write(val)
		case obj.FunctionType:
			f := o.(*obj.CompiledFunction)
			data := make([]byte, 12)
			bin.PutUint32(data, uint32(f.NumParams))
			bin.PutUint32(data[4:], uint32(f.NumLocals))
			bin.PutUint32(data[8:], uint32(len(f.Instructions)))
			buf.Write(data)
			buf.Write(f.Instructions)
		default:
			return fmt.Errorf("unsupported encoding for type %v", o.Type())
		}
	}

	return
}

func decodeObjects(b []byte, n int) (objs []obj.Object, pos int, err error) {
	for ; pos < len(b) && n != 0; n-- {
		t := obj.Type(b[pos])
		pos++

		switch t {
		case obj.NullType:
			objs = append(objs, obj.NullObj)
		case obj.BoolType:
			objs = append(objs, obj.ParseBool(b[pos] == 1))
			pos++
		case obj.IntType:
			objs = append(objs, obj.NewInteger(int64(bin.Uint64(b[pos:]))))
			pos += 8
		case obj.FloatType:
			val := bin.Uint64(b[pos:])
			objs = append(objs, obj.NewFloat(math.Float64frombits(val)))
			pos += 8
		case obj.StringType:
			l := int(bin.Uint32(b[pos:]))
			s := string(b[pos+4 : pos+4+l])
			objs = append(objs, obj.NewString(s))
			pos += 4 + l
		case obj.ErrorType:
			l := int(bin.Uint32(b[pos:]))
			s := string(b[pos+4 : pos+4+l])
			objs = append(objs, obj.NewError(s))
			pos += 4 + l
		case obj.FunctionType:
			numParams := int(bin.Uint32(b[pos:]))
			numLocals := int(bin.Uint32(b[pos+4:]))
			insLen := int(bin.Uint32(b[pos+8:]))
			pos += 12
			ins := code.Instructions(b[pos : pos+insLen])
			objs = append(objs, obj.NewFunctionCompiled(ins, numLocals, numParams, nil))
			pos += insLen
		default:
			return nil, pos, fmt.Errorf("unsupported decoding for type %v", t)
		}
	}

	return
}

func tauEncode(bcode *compiler.Bytecode) ([]byte, error) {
	var buf = new(bytes.Buffer)

	// Write the length of the instructions on the first 4 bytes of the buffer.
	length := make([]byte, 4)
	bin.PutUint32(length, uint32(len(bcode.Instructions)))
	buf.Write(length)
	buf.Write(bcode.Instructions)

	// Write the encoded constants on the tail of the buffer.
	if err := encodeObjects(buf, bcode.Constants); err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func tauDecode(b []byte) (*compiler.Bytecode, error) {
	var bcode = new(compiler.Bytecode)

	if len(b) == 0 {
		return nil, errors.New("decode: empty bytecode")
	}

	// Read the first 4 bytes to determine the instructions length.
	ilen := bin.Uint32(b)
	bcode.Instructions = code.Instructions(b[4 : 4+ilen])

	consts, _, err := decodeObjects(b[4+ilen:], -1)
	if err != nil {
		return nil, err
	}

	bcode.Constants = consts
	return bcode, nil
}
