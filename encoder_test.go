package tau

import (
	"testing"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
	"github.com/NicoNex/tau/internal/tauerr"
)

var (
	bcode = &compiler.Bytecode{
		Instructions: code.Instructions{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		Constants: []obj.Object{
			obj.NewInteger(123),
			obj.String("testing Tau encoding"),
			obj.True,
			obj.NewFloat(123.666),
		},
		Bookmarks: []tauerr.Bookmark{
			{
				Offset: 123,
				LineNo: 10,
				Pos:    3,
				Line:   "this is a bookmark test",
			},
		},
	}
)

func testObjects(t *testing.T, expected, decoded []obj.Object) {
	t.Helper()

	if len(expected) != len(decoded) {
		t.Fatalf("consts length mistmatch, expected %d, got %d", len(expected), len(decoded))
	}

	for i, o := range expected {
		target := decoded[i]

		switch o := o.(type) {
		case obj.Integer:
			i, ok := target.(obj.Integer)
			if !ok {
				t.Fatalf("expected type %v, got %v", o.Type(), target.Type())
			}
			if o.Val() != i.Val() {
				t.Fatalf("value mismatch, expected %v, got %v", o.Val(), i.Val())
			}
		case obj.String:
			s, ok := target.(obj.String)
			if !ok {
				t.Fatalf("expected type %v, got %v", o.Type(), target.Type())
			}
			if o.Val() != s.Val() {
				t.Fatalf("value mismatch, expected %v, got %v", o.Val(), s.Val())
			}
		case *obj.Boolean:
			b, ok := target.(*obj.Boolean)
			if !ok {
				t.Fatalf("expected type %v, got %v", o.Type(), target.Type())
			}
			if b != o {
				t.Fatalf("value mismatch, expected %v, got %v", o, b)
			}
		case obj.Float:
			f, ok := target.(obj.Float)
			if !ok {
				t.Fatalf("expected type %v, got %v", o.Type(), target.Type())
			}
			if f.Val() != o.Val() {
				t.Fatalf("value mismatch, expected %v, got %v", o.Val(), f.Val())
			}
		}
	}
}

func TestEncode(t *testing.T) {
	var pos int

	b, err := tauEncode(bcode)
	if err != nil {
		t.Fatal(err)
	}

	ilen := int(bin.Uint32(b))
	if ilen != len(bcode.Instructions) {
		t.Fatalf("instruction length mismatch, expected %d, got %d", len(bcode.Instructions), ilen)
	}
	pos += 4

	generatedBytecode := code.Instructions(b[pos : pos+ilen])
	for i, inst := range generatedBytecode {
		if inst != bcode.Instructions[i] {
			t.Fatalf(
				"wrong instruction at position %d, expected %d, got %d",
				pos+i,
				bcode.Instructions[i],
				generatedBytecode[i],
			)
		}
	}
	pos += ilen

	clen := int(bin.Uint32(b[pos:]))
	objs, p, err := decodeObjects(b[pos+4:], clen)
	pos += 4 + p
	testObjects(t, bcode.Constants, objs)

	blen := int(bin.Uint32(b[pos:]))
	bmarks, _ := decodeBookmarks(b[pos+4:], blen)

	if len(bmarks) != len(bcode.Bookmarks) {
		t.Fatalf("bookmarks length mistmatch, expected %d, got %d", len(bcode.Bookmarks), len(bmarks))
	}

	for i, b := range bmarks {
		if b != bcode.Bookmarks[i] {
			t.Fatalf("bookmark mismatch at index %d", i)
		}
	}
}
