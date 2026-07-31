package doc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NicoNex/tau/internal/mod"
)

const src = `# mod - a module written to be read.
#
# The second paragraph of what the module is about.

other = import("other")

# Answer is the number.
Answer = 42

# hidden is not given away.
hidden = fn() { 1 }

# Counter builds a thing that counts.
#
#	c = mod.Counter(0)
#	c.Inc()
Counter = fn(start, step) {
	c = new()
	c.n = start # a comment ending the line, which eats its newline

	# Inc adds one and returns the total.
	c.Inc = fn() {
		c.n = c.n + step
		return c.n
	}

	# Nested holds a thing of its own.
	c.Nested = fn() {
		d = new()

		# Deep is as far down as this goes.
		d.Deep = fn() { 1 }
		return d
	}

	# Trailing comes after a line that ends in a comment.
	c.Trailing = fn() { 2 }

	# secret stays in.
	c.secret = fn() { 0 }
	return c
}

# Loose is a comment with a blank line under it.

Loose = 1
`

func load(t *testing.T) Package {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "mod.tau")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	p, err := Load("mod", path)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestExported(t *testing.T) {
	p := load(t)

	var names []string
	for _, e := range p.Entries {
		names = append(names, e.Name)
	}

	want := "Answer Counter Loose"
	if got := strings.Join(names, " "); got != want {
		t.Errorf("top level names are %q, want %q", got, want)
	}
}

func TestPackageDoc(t *testing.T) {
	p := load(t)

	if !strings.HasPrefix(p.Doc, "mod - a module written to be read.") {
		t.Errorf("package doc is %q", p.Doc)
	}
	if !strings.Contains(p.Doc, "second paragraph") {
		t.Error("the package doc stops at the first paragraph")
	}
}

func TestSignatureAndValue(t *testing.T) {
	p := load(t)

	c, ok := p.Find("Counter")
	if !ok {
		t.Fatal("no Counter")
	}
	if c.Sig != "fn(start, step)" {
		t.Errorf("signature is %q, want %q", c.Sig, "fn(start, step)")
	}

	a, _ := p.Find("Answer")
	if a.Val != "42" {
		t.Errorf("Answer holds %q, want %q", a.Val, "42")
	}
}

// TestDescent is the point of the package: what Go documents as methods is
// here what a constructor puts on the object it returns, and it goes as deep
// as it is asked to.
func TestDescent(t *testing.T) {
	p := load(t)

	for _, tt := range []struct{ path, doc string }{
		{"Counter.Inc", "Inc adds one and returns the total."},
		{"Counter.Nested", "Nested holds a thing of its own."},
		{"Counter.Nested.Deep", "Deep is as far down as this goes."},
		// The lexer gives no newline after a comment, so a statement ending
		// in one used to swallow whatever followed it.
		{"Counter.Trailing", "Trailing comes after a line that ends in a comment."},
	} {
		e, ok := p.Find(tt.path)
		if !ok {
			t.Errorf("%s is not there", tt.path)
			continue
		}
		if e.Doc != tt.doc {
			t.Errorf("%s says %q, want %q", tt.path, e.Doc, tt.doc)
		}
	}

	if _, ok := p.Find("Counter.secret"); ok {
		t.Error("a lower case field was given away")
	}
	if _, ok := p.Find("Counter.n"); ok {
		t.Error("a lower case field was given away")
	}
}

// TestBlankLine is what tells a comment about a name from one that happens to
// sit above it.
func TestBlankLine(t *testing.T) {
	p := load(t)

	e, ok := p.Find("Loose")
	if !ok {
		t.Fatal("no Loose")
	}
	if e.Doc != "" {
		t.Errorf("a comment across an empty line was taken as a doc: %q", e.Doc)
	}
}

func TestFindMissing(t *testing.T) {
	p := load(t)

	if _, ok := p.Find("Nope"); ok {
		t.Error("found a name that isn't there")
	}
	if _, ok := p.Find("Counter.Nope"); ok {
		t.Error("found a field that isn't there")
	}
}

// TestStdlib reads every module of the standard library: they are what this
// was written for, and the shape it expects is the one they are written in.
func TestStdlib(t *testing.T) {
	entries, err := os.ReadDir("../../stdlib")
	if err != nil {
		t.Skip("no stdlib to read")
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		// A directory holding only other modules (crypto, encoding) is not
		// one itself: there is nothing there to read.
		dir := filepath.Join("../../stdlib", e.Name())
		if !mod.IsDirModule(dir) {
			continue
		}

		p, err := Load(e.Name(), dir)
		if err != nil {
			t.Errorf("%s: %v", e.Name(), err)
			continue
		}
		if len(p.Entries) == 0 {
			t.Errorf("%s: no exported name found", e.Name())
		}
	}
}
