package main

import (
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
)

// position is an LSP position: a zero based line and a character counted in
// UTF-16 code units, which is what the protocol means by "character" no
// matter what the file is encoded in.
type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type textRange struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

// document is an open buffer together with the byte offset of each line, so
// that a position can be turned into an offset and back without rescanning
// the text.
type document struct {
	uri     string
	path    string
	version int
	text    string

	lines []int // byte offset of the start of each line
	syms  *fileInfo
}

func newDocument(uri, text string, version int) *document {
	d := &document{uri: uri, path: uriToPath(uri), version: version}
	d.setText(text)
	return d
}

func (d *document) setText(text string) {
	d.text = text
	d.lines = d.lines[:0]
	d.lines = append(d.lines, 0)
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			d.lines = append(d.lines, i+1)
		}
	}
	d.syms = nil
}

// info analyses the buffer once and caches the result until it changes.
func (d *document) info() *fileInfo {
	if d.syms == nil {
		d.syms = analyse(d.path, d.text)
	}
	return d.syms
}

// lineRange returns the byte bounds of line n, newline excluded.
func (d *document) lineRange(n int) (start, end int) {
	if n < 0 || n >= len(d.lines) {
		return len(d.text), len(d.text)
	}
	start = d.lines[n]
	end = len(d.text)
	if n+1 < len(d.lines) {
		end = d.lines[n+1] - 1
	}
	if end > 0 && end <= len(d.text) && end > start && d.text[end-1] == '\r' {
		end--
	}
	return start, end
}

func (d *document) line(n int) string {
	start, end := d.lineRange(n)
	return d.text[start:end]
}

// lineOf returns the zero based line holding the given byte offset.
func (d *document) lineOf(offset int) int {
	lo, hi := 0, len(d.lines)-1
	for lo < hi {
		mid := (lo + hi + 1) / 2
		if d.lines[mid] <= offset {
			lo = mid
		} else {
			hi = mid - 1
		}
	}
	return lo
}

// position converts a byte offset into an LSP position.
func (d *document) position(offset int) position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(d.text) {
		offset = len(d.text)
	}
	line := d.lineOf(offset)
	return position{Line: line, Character: utf16Len(d.text[d.lines[line]:offset])}
}

// offset converts an LSP position into a byte offset, clamped to the line it
// names so that a stale position from the client cannot walk off the buffer.
func (d *document) offset(p position) int {
	if p.Line < 0 {
		return 0
	}
	if p.Line >= len(d.lines) {
		return len(d.text)
	}
	start, end := d.lineRange(p.Line)
	line := d.text[start:end]

	var units int
	for i, r := range line {
		if units >= p.Character {
			return start + i
		}
		units += runeUTF16Len(r)
	}
	return end
}

func (d *document) rangeOf(start, end int) textRange {
	return textRange{Start: d.position(start), End: d.position(end)}
}

func utf16Len(s string) int {
	var n int
	for _, r := range s {
		n += runeUTF16Len(r)
	}
	return n
}

// runeUTF16Len is how many UTF-16 code units a rune takes: one, or two for
// anything outside the basic plane, which is what the protocol counts.
func runeUTF16Len(r rune) int {
	if r > 0xFFFF {
		return 2
	}
	return 1
}

// uriToPath turns a file:// URI into a filesystem path. Anything that is not
// a file URI comes back as it is: the server still needs a stable key for it,
// it just cannot read it from disk.
func uriToPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return uri
	}
	p := u.Path
	if runtime.GOOS == "windows" {
		p = strings.TrimPrefix(p, "/")
		p = filepath.FromSlash(p)
	}
	return p
}

func pathToURI(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	abs = filepath.ToSlash(abs)
	if !strings.HasPrefix(abs, "/") {
		abs = "/" + abs
	}
	u := url.URL{Scheme: "file", Path: abs}
	return u.String()
}
