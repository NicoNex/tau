package bundle

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

// A bundle is written in a shape a C reader can walk without a library, so
// that the runtime a bundled program is built on needs no Go in it at all.
// Everything is a big endian uint32 followed by what it counts, and the
// modules come in the order they have to be loaded, which is the order they
// were compiled for:
//
//	uint32 nmodules
//	  for each: name, bytecode, uint32 nexports, for each: name, uint32 index
//	uint32 nplugins
//	  for each: name, the shared object
//	the bytecode of the program
//
// where a name and a blob are both a uint32 length followed by that many
// bytes. The bytecode of the program comes last and takes what is left, so it
// carries no length of its own.

// Encode writes the bundle in the shape Decode and the C reader expect.
func (b Bundle) Encode() []byte {
	var buf []byte

	buf = binary.BigEndian.AppendUint32(buf, uint32(len(b.Order)))
	for _, name := range b.Order {
		m := b.Modules[name]

		buf = appendBlob(buf, []byte(name))
		buf = appendBlob(buf, m.Bytecode)
		buf = binary.BigEndian.AppendUint32(buf, uint32(len(m.Exports)))

		// Sorted, so that the same program bundled twice gives the same file.
		for _, exp := range sortedKeys(m.Exports) {
			buf = appendBlob(buf, []byte(exp))
			buf = binary.BigEndian.AppendUint32(buf, uint32(m.Exports[exp]))
		}
	}

	buf = binary.BigEndian.AppendUint32(buf, uint32(len(b.Plugins)))
	for _, name := range sortedKeys(b.Plugins) {
		buf = appendBlob(buf, []byte(name))
		buf = appendBlob(buf, b.Plugins[name])
	}

	return append(buf, b.Bytecode...)
}

// Decode reads back what Encode wrote.
func (b *Bundle) decode(raw []byte) error {
	r := &reader{buf: raw}

	nmods := r.uint32()
	b.Order = make([]string, 0, nmods)
	b.Modules = make(map[string]ModuleCode, nmods)

	for i := uint32(0); i < nmods; i++ {
		name := string(r.blob())
		m := ModuleCode{Bytecode: r.blob(), Exports: map[string]int{}}

		for n := r.uint32(); n > 0; n-- {
			exp := string(r.blob())
			m.Exports[exp] = int(r.uint32())
		}

		b.Order = append(b.Order, name)
		b.Modules[name] = m
	}

	nplugins := r.uint32()
	b.Plugins = make(map[string][]byte, nplugins)
	for i := uint32(0); i < nplugins; i++ {
		name := string(r.blob())
		b.Plugins[name] = r.blob()
	}

	if r.err != nil {
		return r.err
	}
	b.Bytecode = r.rest()
	if len(b.Bytecode) == 0 {
		return errors.New("bundle: no program in it")
	}
	return nil
}

func appendBlob(buf, b []byte) []byte {
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(b)))
	return append(buf, b...)
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// reader walks the bundle and remembers the first thing that went wrong, so
// that a truncated file is reported once rather than at every field.
type reader struct {
	buf []byte
	pos int
	err error
}

func (r *reader) take(n int) []byte {
	if r.err != nil {
		return nil
	}
	if n < 0 || r.pos+n > len(r.buf) {
		r.err = fmt.Errorf("bundle: it ends after %d bytes, in the middle of something", len(r.buf))
		return nil
	}

	b := r.buf[r.pos : r.pos+n]
	r.pos += n
	return b
}

func (r *reader) uint32() uint32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return binary.BigEndian.Uint32(b)
}

func (r *reader) blob() []byte { return r.take(int(r.uint32())) }

func (r *reader) rest() []byte {
	if r.err != nil {
		return nil
	}
	return r.buf[r.pos:]
}
