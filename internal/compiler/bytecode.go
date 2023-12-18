package compiler

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/tauerr"
)

type Bytecode struct {
	Instructions code.Instructions
	Constants    []obj.Object
	Bookmarks    []tauerr.Bookmark
	NumDefs      int
}

type encoder struct {
	bytes.Buffer
}

func (e *encoder) Write(b []byte) (err error) {
	_, err = e.Buffer.Write(b)
	return
}

func (e *encoder) WriteString(s string) (err error) {
	_, err = e.Buffer.WriteString(s)
	return
}

func (e *encoder) WriteValue(data any) error {
	return binary.Write(&e.Buffer, binary.BigEndian, data)
}

func (e *encoder) WriteBookmarks(bookmarks ...tauerr.Bookmark) (err error) {
	for _, b := range bookmarks {
		err = errors.Join(err, e.WriteValue(b.Offset))
		err = errors.Join(err, e.WriteValue(b.LineNo))
		err = errors.Join(err, e.WriteValue(b.Pos))
		err = errors.Join(err, e.WriteValue(len(b.Line)))
		err = errors.Join(err, e.WriteString(b.Line))
	}
	return
}

func (e *encoder) WriteObjects(objs ...obj.Object) (err error) {
	for _, object := range objs {
		err = errors.Join(err, e.WriteByte(byte(object.Type())))

		switch o := object.(type) {
		case obj.Null:
			break
		case obj.Boolean:
			err = errors.Join(err, e.WriteValue(o.Val()))
		case obj.Integer:
			err = errors.Join(err, e.WriteValue(o.Val()))
		case obj.Float:
			err = errors.Join(err, e.WriteValue(o.Val()))
		case obj.String:
			err = errors.Join(err, e.WriteValue(o.Val()))
		case obj.CompiledFunction:
			err = errors.Join(err, e.WriteValue(o.NumParams))
			err = errors.Join(err, e.WriteValue(o.NumLocals))
			err = errors.Join(err, e.WriteValue(len(o.Instructions)))
			err = errors.Join(err, e.Write(o.Instructions))
			err = errors.Join(err, e.WriteValue(len(o.Bookmarks)))
			err = errors.Join(err, e.WriteBookmarks(o.Bookmarks...))
		}
	}
	return
}

func (b Bytecode) Encode() (data []byte, err error) {
	var e encoder

	err = errors.Join(err, e.WriteValue(b.NumDefs))
	err = errors.Join(err, e.WriteValue(len(b.Instructions)))
	err = errors.Join(err, e.Write(b.Instructions))
	err = errors.Join(err, e.WriteValue(len(b.Constants)))
	err = errors.Join(err, e.WriteObjects(b.Constants...))
	err = errors.Join(err, e.WriteValue(len(b.Bookmarks)))
	err = errors.Join(err, e.WriteBookmarks(b.Bookmarks...))
	return e.Bytes(), err
}

type decoder struct {
	pos int
	b   []byte
}

func (d *decoder) Byte() (b byte) {
	b = d.b[d.pos]
	d.pos++
	return
}

func (d *decoder) Int() (i int) {
	i = int(binary.BigEndian.Uint32(d.b[d.pos:]))
	d.pos += 4
	return
}

func (d *decoder) Uint64() (i uint64) {
	i = binary.BigEndian.Uint64(d.b[d.pos:])
	d.pos += 8
	return
}

func (d *decoder) Int64() (i int64) {
	return int64(d.Uint64())
}

func (d *decoder) Float64() (f float64) {
	return math.Float64frombits(d.Uint64())
}

func (d *decoder) Bytes(len int) (b []byte) {
	b = d.b[d.pos : d.pos+len]
	d.pos += len
	return
}

func (d *decoder) String(len int) (s string) {
	s = string(d.b[d.pos : d.pos+len])
	d.pos += len
	return
}

func (d *decoder) Objects(len int) (o []obj.Object) {
	for i := 0; i < len; i++ {
		switch t := obj.Type(d.Byte()); t {
		case obj.NullType:
			o = append(o, obj.NullObj)
		case obj.BoolType:
			o = append(o, obj.ParseBool(d.Byte() == 1))
		case obj.IntType:
			o = append(o, obj.NewInteger(d.Int64()))
		case obj.FloatType:
			o = append(o, obj.NewFloat(d.Float64()))
		case obj.StringType:
			o = append(o, obj.NewString(d.String(d.Int())))
		case obj.FunctionType:
			// The order of the fields has to reflect the data layout in the encoded bytecode.
			o = append(o, &obj.CompiledFunction{
				NumParams:    d.Int(),
				NumLocals:    d.Int(),
				Instructions: d.Bytes(d.Int()),
				Bookmarks:    d.Bookmarks(d.Int()),
			})
		default:
			panic(fmt.Sprintf("decoder: unsupported decoding for type %s", t))
		}
	}
	return
}

func (d *decoder) Bookmarks(len int) (b []tauerr.Bookmark) {
	for i := 0; i < len; i++ {
		b = append(b, tauerr.Bookmark{
			Offset: d.Int(),
			LineNo: d.Int(),
			Pos:    d.Int(),
			Line:   d.String(d.Int()),
		})
	}
	return
}

func Decode(b []byte) *Bytecode {
	var d = decoder{b: b}

	// The order of the fields has to reflect the data layout in the encoded bytecode.
	return &Bytecode{
		NumDefs:      d.Int(),
		Instructions: d.Bytes(d.Int()),
		Constants:    d.Objects(d.Int()),
		Bookmarks:    d.Bookmarks(d.Int()),
	}
}
