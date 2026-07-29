package tau

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
var execMagic = []byte("TAUX")

// The magic is four bytes and the length that follows it is eight.
const execTrailerLen = 4 + 8

// errNoPayload says the executable carries no program, which is the ordinary
// case for the interpreter itself.
var errNoPayload = errors.New("no bundle appended")

// payloadAt says where the appended bundle of a file this long begins and how
// long it is, reading the trailer it was handed from the end of that file. It
// reports false for a file that carries nothing, which is what an interpreter
// on its own looks like.
func payloadAt(trailer []byte, size int64) (start, length int64, ok bool) {
	if len(trailer) != execTrailerLen || !bytes.Equal(trailer[:len(execMagic)], execMagic) {
		return 0, 0, false
	}

	length = int64(binary.BigEndian.Uint64(trailer[len(execMagic):]))
	start = size - int64(execTrailerLen) - length
	if length <= 0 || start < 0 {
		return 0, 0, false
	}
	return start, length, true
}

// embeddedBundle returns the bundle appended to the running executable.
func embeddedBundle() ([]byte, error) {
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
	if size < int64(execTrailerLen) {
		return nil, errNoPayload
	}

	trailer := make([]byte, execTrailerLen)
	if _, err := f.ReadAt(trailer, size-int64(execTrailerLen)); err != nil {
		return nil, err
	}

	start, length, ok := payloadAt(trailer, size)
	if !ok {
		return nil, errNoPayload
	}

	raw := make([]byte, length)
	if _, err := f.ReadAt(raw, start); err != nil {
		return nil, err
	}
	return raw, nil
}

// HasEmbeddedProgram reports whether this executable carries a program of its
// own, in which case it is that program and not the interpreter command.
func HasEmbeddedProgram() bool {
	_, err := embeddedBundle()
	return err == nil
}

// RunEmbedded runs the program appended to this executable.
func RunEmbedded() error {
	raw, err := embeddedBundle()
	if err != nil {
		return err
	}

	bytecode, clean, err := openBundle(raw)
	if err != nil {
		return err
	}
	defer clean()

	self, err := os.Executable()
	if err != nil {
		self = os.Args[0]
	}
	return runBytecode(self, bytecode)
}

// BuildExecutable writes path and everything it imports into a copy of this
// interpreter, so that the result runs on its own.
func BuildExecutable(path, out string) error {
	b, err := Bundle(path)
	if err != nil {
		return err
	}

	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("bundle: cannot find the interpreter to copy: %w", err)
	}
	// The interpreter as it is now, without anything an earlier bundle may
	// have left on it: building from an executable that already carries a
	// program would otherwise stack one program on top of another.
	interp, err := interpreterOnly(self)
	if err != nil {
		return err
	}

	if out == "" {
		out = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if runtime.GOOS == "windows" {
			out += ".exe"
		}
	}

	var buf bytes.Buffer
	buf.Write(interp)
	buf.Write(b)
	buf.Write(execMagic)
	binary.Write(&buf, binary.BigEndian, uint64(len(b)))

	if err := os.WriteFile(out, buf.Bytes(), 0755); err != nil {
		return err
	}

	if runtime.GOOS == "darwin" {
		fmt.Fprintf(
			os.Stderr,
			"bundle: %s is written; on macOS run `codesign -s - %s` before it will start, "+
				"appending to an executable invalidates the signature it came with\n",
			out, out,
		)
	}
	return nil
}

// interpreterOnly reads an executable without the bundle it may be carrying.
func interpreterOnly(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if len(raw) < execTrailerLen {
		return raw, nil
	}

	end, _, ok := payloadAt(raw[len(raw)-execTrailerLen:], int64(len(raw)))
	if !ok {
		return raw, nil
	}
	return raw[:end], nil
}
