package tau

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// appended builds what BuildExecutable writes: an interpreter, a bundle, and
// the trailer that says how long the bundle is.
func appended(interp, payload []byte) []byte {
	var buf bytes.Buffer

	buf.Write(interp)
	buf.Write(payload)
	buf.Write(execMagic)
	binary.Write(&buf, binary.BigEndian, uint64(len(payload)))
	return buf.Bytes()
}

func TestPayloadAt(t *testing.T) {
	interp := []byte("this stands for the interpreter")
	payload := []byte("this stands for the bundle")
	file := appended(interp, payload)

	start, length, ok := payloadAt(file[len(file)-execTrailerLen:], int64(len(file)))
	if !ok {
		t.Fatal("a file with a bundle appended reported none")
	}
	if got := file[start : start+length]; !bytes.Equal(got, payload) {
		t.Fatalf("read back %q, want %q", got, payload)
	}
	if start != int64(len(interp)) {
		t.Fatalf("the bundle starts at %d, want %d", start, len(interp))
	}
}

func TestPayloadAtWithoutOne(t *testing.T) {
	// An interpreter on its own, which is the ordinary case and must not be
	// mistaken for one carrying a program.
	plain := []byte("an interpreter and nothing after it")

	if _, _, ok := payloadAt(plain[len(plain)-execTrailerLen:], int64(len(plain))); ok {
		t.Fatal("a plain interpreter reported a bundle")
	}
}

func TestPayloadAtRefusesNonsense(t *testing.T) {
	interp := []byte("interpreter")
	file := appended(interp, []byte("bundle"))

	// A length longer than the file itself would send the reader before its
	// start: it has to be refused rather than trusted.
	binary.BigEndian.PutUint64(file[len(file)-8:], uint64(len(file)+1))
	if _, _, ok := payloadAt(file[len(file)-execTrailerLen:], int64(len(file))); ok {
		t.Fatal("a length past the start of the file was accepted")
	}

	// And a length of zero says there is nothing there.
	binary.BigEndian.PutUint64(file[len(file)-8:], 0)
	if _, _, ok := payloadAt(file[len(file)-execTrailerLen:], int64(len(file))); ok {
		t.Fatal("an empty bundle was accepted")
	}
}

func TestBundleIsStrippedBeforeAppending(t *testing.T) {
	// Bundling from an interpreter that already carries a program has to start
	// from the interpreter, otherwise every build would stack one program on
	// top of the last.
	interp := []byte("this stands for the interpreter")
	once := appended(interp, []byte("the first program"))
	twice := appended(once, []byte("the second program"))

	end, _, ok := payloadAt(twice[len(twice)-execTrailerLen:], int64(len(twice)))
	if !ok {
		t.Fatal("the outer bundle was not found")
	}
	stripped := twice[:end]

	end, _, ok = payloadAt(stripped[len(stripped)-execTrailerLen:], int64(len(stripped)))
	if !ok {
		t.Fatal("the inner bundle was not found")
	}
	if got := stripped[:end]; !bytes.Equal(got, interp) {
		t.Fatalf("stripping twice left %q, want %q", got, interp)
	}
}
