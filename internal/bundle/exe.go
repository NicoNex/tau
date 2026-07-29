package bundle

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// A bundled program can be carried by a copy of the interpreter itself, which
// makes a single file that runs where nothing is installed. The bundle is put
// after the end of the executable and a trailer at the very end says how long
// it is:
//
//	[ the interpreter ][ the bundle ][ "TAUX" | uint64 length ]
//
// The trailer goes last because that is the only place that can be found
// without knowing anything about the format of the executable in front of it:
// seek to the end, step back, and look. An interpreter with nothing appended
// has no trailer and goes on behaving like the command it is.
//
// Nothing is linked and no compiler is needed: a loader reads an executable
// through its own headers and never looks at what follows them, so appending
// leaves it runnable. That also means an interpreter downloaded ready-made can
// make executables, not only one built here.
// ExecMagic marks an executable carrying a program.
var ExecMagic = []byte("TAUX")

// The magic is four bytes and the length that follows it is eight.
// TrailerLen is the magic plus the eight bytes of length after it.
const TrailerLen = 4 + 8

// ErrNoPayload says the executable carries no program, which is the ordinary
// case for the interpreter itself.
var ErrNoPayload = errors.New("no bundle appended")

// payloadAt says where the appended bundle of a file this long begins and how
// long it is, reading the trailer it was handed from the end of that file. It
// reports false for a file that carries nothing, which is what an interpreter
// on its own looks like.
// PayloadAt says where the appended bundle of a file this long begins and how
// long it is, reading the trailer it was handed from the end of that file. It
// reports false for a file that carries nothing, which is what an interpreter
// on its own looks like.
func PayloadAt(trailer []byte, size int64) (start, length int64, ok bool) {
	if len(trailer) != TrailerLen || !bytes.Equal(trailer[:len(ExecMagic)], ExecMagic) {
		return 0, 0, false
	}

	length = int64(binary.BigEndian.Uint64(trailer[len(ExecMagic):]))
	start = size - int64(TrailerLen) - length
	if length <= 0 || start < 0 {
		return 0, 0, false
	}
	return start, length, true
}

// embeddedBundle returns the bundle appended to the running executable.
func Embedded() ([]byte, error) {
	self, err := os.Executable()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(self)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	size, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, err
	}
	if size < int64(TrailerLen) {
		return nil, ErrNoPayload
	}

	trailer := make([]byte, TrailerLen)
	if _, err := f.ReadAt(trailer, size-int64(TrailerLen)); err != nil {
		return nil, err
	}

	start, length, ok := PayloadAt(trailer, size)
	if !ok {
		return nil, ErrNoPayload
	}

	raw := make([]byte, length)
	if _, err := f.ReadAt(raw, start); err != nil {
		return nil, err
	}
	return raw, nil
}

// HasEmbeddedProgram reports whether this executable carries a program of its
// own, in which case it is that program and not the interpreter command.
func HasEmbedded() bool {
	_, err := Embedded()
	return err == nil
}

// RunEmbedded runs the program appended to this executable.
func RunEmbedded() error {
	raw, err := Embedded()
	if err != nil {
		return err
	}

	bytecode, clean, err := Open(raw)
	if err != nil {
		return err
	}
	defer clean()

	self, err := os.Executable()
	if err != nil {
		self = os.Args[0]
	}
	return Run(self, bytecode)
}


