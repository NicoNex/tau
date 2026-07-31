package doc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLines is what the link to the source rests on: the line an entry says
// it was read from has to be the line the name is written on. Nothing else
// would notice if it were off by one - the page would open, on the wrong
// line - so it is checked here.
func TestLines(t *testing.T) {
	for _, mod := range []string{"sync", "sync/atomic", "buffer", "strings", "os"} {
		p, err := Load(mod, filepath.Join("../../stdlib", mod))
		if err != nil {
			t.Fatalf("%s: %v", mod, err)
		}

		var walk func([]Entry)
		walk = func(entries []Entry) {
			for _, e := range entries {
				src, err := os.ReadFile(e.File)
				if err != nil {
					t.Fatalf("%s: %v", e.Name, err)
				}

				lines := strings.Split(string(src), "\n")
				if e.Line < 1 || e.Line > len(lines) {
					t.Errorf("%s: %s says line %d of %d", mod, e.Name, e.Line, len(lines))
					continue
				}
				if line := lines[e.Line-1]; !strings.Contains(line, e.Name) {
					t.Errorf("%s: %s says line %d, which is %q", mod, e.Name, e.Line, line)
				}
				walk(e.Children)
			}
		}
		walk(p.Entries)
	}
}

// TestHighlightLines is the other half: a listing has one row per line of the
// file, and the colouring never leaves a line holding half a span.
func TestHighlightLines(t *testing.T) {
	src, err := os.ReadFile("../../stdlib/sync/sync.tau")
	if err != nil {
		t.Skip("no stdlib to read")
	}

	want := strings.Split(string(src), "\n")
	got := highlightLines(string(src))

	if len(got) != len(want) {
		t.Fatalf("%d lines of source became %d rows", len(want), len(got))
	}

	for i, row := range got {
		s := string(row)
		if strings.Count(s, "<span") != strings.Count(s, "</span>") {
			t.Errorf("line %d has an unclosed span: %s", i+1, s)
		}
	}
}
