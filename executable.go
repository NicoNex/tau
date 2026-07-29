package tau

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	bundlepkg "github.com/NicoNex/tau/internal/bundle"
)

// HasEmbeddedProgram reports whether this executable carries a program of its
// own, in which case it is that program and not the interpreter command.
func HasEmbeddedProgram() bool { return bundlepkg.HasEmbedded() }

// RunEmbedded runs the program appended to this executable.
func RunEmbedded() error { return bundlepkg.RunEmbedded() }

// runtimeStub is the smallest thing a bundled program can be built on: a
// runtime with the VM and the objects but without the lexer, the parser, the
// syntax tree and the compiler, which a compiled program never asks for. It
// sits next to the interpreter or where make install put it. Falling back to
// the interpreter itself keeps bundling working for a tau on its own, at the
// cost of carrying a compiler the program will not use.
func runtimeStub() (string, bool) {
	name := "tau-rt"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}

	var dirs []string
	if self, err := os.Executable(); err == nil {
		// Next to the interpreter, and in the lib directory beside its bin
		// one: that is where make install puts it, PREFIX/bin/tau and
		// PREFIX/lib/tau/tau-rt.
		dir := filepath.Dir(self)
		dirs = append(dirs, dir, filepath.Join(dir, "..", "lib", "tau"))
	}
	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "lib", "tau"))
	}
	dirs = append(dirs, "/usr/local/lib/tau", "/lib/tau")

	for _, d := range dirs {
		p := filepath.Join(d, name)
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p, true
		}
	}
	return "", false
}

// BuildExecutable writes path and everything it imports into a copy of this
// interpreter, so that the result runs on its own.
func BuildExecutable(path, out string) error {
	b, err := Bundle(path)
	if err != nil {
		return err
	}

	base, ok := runtimeStub()
	if !ok {
		if base, err = os.Executable(); err != nil {
			return fmt.Errorf("bundle: cannot find anything to build on: %w", err)
		}
	}
	// Without anything an earlier bundle may have left on it: building from an
	// executable that already carries a program would otherwise stack one
	// program on top of another.
	interp, err := interpreterOnly(base)
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
	buf.Write(bundlepkg.ExecMagic)
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

	if len(raw) < bundlepkg.TrailerLen {
		return raw, nil
	}

	end, _, ok := bundlepkg.PayloadAt(raw[len(raw)-bundlepkg.TrailerLen:], int64(len(raw)))
	if !ok {
		return raw, nil
	}
	return raw[:end], nil
}
