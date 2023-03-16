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
	"github.com/NicoNex/tau/internal/tauerr"
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
			encodeBookmarks(buf, f.Bookmarks)
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
			pos += insLen

			nbmark := int(bin.Uint32(b[pos:]))
			bmarks, bpos := decodeBookmarks(b[pos+4:], nbmark)
			objs = append(objs, obj.NewFunctionCompiled(ins, numLocals, numParams, bmarks))
			pos += 4 + bpos
		default:
			return nil, pos, fmt.Errorf("unsupported decoding for type %v", t)
		}
	}

	return
}

func decodeBookmarks(b []byte, n int) (bmarks []tauerr.Bookmark, pos int) {
	for ; pos < len(b) && n > 0; n-- {
		offset := btoi(b[pos:])
		lineno := btoi(b[pos+4:])
		bpos := btoi(b[pos+8:])
		slen := btoi(b[pos+12:])
		line := string(b[pos+16 : pos+16+slen])
		bmarks = append(bmarks, tauerr.Bookmark{
			Offset: offset,
			LineNo: lineno,
			Pos:    bpos,
			Line:   line,
		})
		pos += 16 + slen
	}
	return
}

func encodeBookmarks(buf *bytes.Buffer, bmarks []tauerr.Bookmark) {
	buf.Write(itob(len(bmarks)))

	for _, b := range bmarks {
		data := make([]byte, 16)
		bin.PutUint32(data, uint32(b.Offset))
		bin.PutUint32(data[4:], uint32(b.LineNo))
		bin.PutUint32(data[8:], uint32(b.Pos))
		bin.PutUint32(data[12:], uint32(len(b.Line)))
		buf.Write(data)
		buf.WriteString(b.Line)
	}
}

func itob(i int) (b []byte) {
	b = make([]byte, 4)
	bin.PutUint32(b, uint32(i))
	return
}

func btoi(b []byte) int {
	return int(bin.Uint32(b))
}

func tauEncode(bcode *compiler.Bytecode) ([]byte, error) {
	var buf = new(bytes.Buffer)

	// Write the number of definitions to the first 4 bytes of the buffer.
	buf.Write(itob(bcode.NumDefs))

	// Write the length of the instructions to the following 4 bytes of the buffer.
	buf.Write(itob(len(bcode.Instructions)))
	buf.Write(bcode.Instructions)

	// Write the length and the encoded constants to the tail of the buffer.
	buf.Write(itob(len(bcode.Constants)))
	if err := encodeObjects(buf, bcode.Constants); err != nil {
		return []byte{}, err
	}
	// Write the encoded bookmarks to the tail of the buffer.
	encodeBookmarks(buf, bcode.Bookmarks)
	return buf.Bytes(), nil
}

func tauDecode(b []byte) (*compiler.Bytecode, error) {
	if len(b) == 0 {
		return nil, errors.New("decode: empty bytecode")
	}

	var (
		pos   int
		bcode = new(compiler.Bytecode)
	)

	bcode.NumDefs = btoi(b)
	pos += 4

	// Read the first 4 bytes to determine the instructions length.
	ilen := btoi(b[pos:])
	pos += 4
	bcode.Instructions = code.Instructions(b[pos : pos+ilen])
	pos += ilen

	clen := btoi(b[pos:])
	pos += 4
	consts, clen, err := decodeObjects(b[pos:], clen)
	if err != nil {
		return nil, err
	}
	bcode.Constants = consts
	pos += clen

	blen := btoi(b[pos:])
	pos += 4
	bmarks, _ := decodeBookmarks(b[pos:], blen)
	bcode.Bookmarks = bmarks
	return bcode, nil
}
