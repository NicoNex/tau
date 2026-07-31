// Package doc reads a module and reports what it gives whoever imports it:
// the names it exports, the comment written above each of them, and the names
// those in turn hold.
//
// Like the formatter, it works on the token stream and not on the AST, for
// the same reason: the AST has no comments, and comments are the whole point
// here. What it looks for is the shape the standard library is written in - a
// name assigned at the top level, with the paragraph explaining it directly
// above.
//
// Tau has no types, so what Go documents as the methods of one is here the
// fields a constructor gives the object it returns:
//
//	Builder = fn() {
//		b = new()
//		# Write appends s to the buffer.
//		b.Write = fn(s) { ... }
//		return b
//	}
//
// Those are found by descending into the body of an exported function and
// reading it by the same rule as the file around it. That descent is not a
// special case for constructors: it repeats for as deep as it is asked to go,
// so a field holding a function of its own is read the same way.
package doc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/NicoNex/tau/internal/item"
	"github.com/NicoNex/tau/internal/lexer"
	"github.com/NicoNex/tau/internal/mod"
)

// Entry is one name a module exports, or one field of the object such a name
// builds.
type Entry struct {
	Name string
	// Sig is "fn(a, b)" for a function, empty for anything else.
	Sig string
	// Val is what a name that isn't a function is assigned, when that fits on
	// a line. A long expression is left out: the source is the place for it.
	Val string
	// Doc is the comment above the name, with the '#' and one space after it
	// taken off, still holding the empty lines that separate its paragraphs.
	Doc      string
	Children []Entry
	File     string
	Line     int

	// returns are the names this could be handing back, in the order they
	// were written. A constructor that ends in another function's result
	// documents what that one builds, since that is what the caller gets.
	returns []string
}

// Package is a module as this package reads it.
type Package struct {
	// Path is the module as it was asked for, which is what an import needs.
	Path string
	// Dir is empty for a module that is a single file.
	Dir     string
	Files   []string
	Doc     string
	Entries []Entry
}

// Var is the name an import of this module is usually given: the last part of
// its path, which is what the standard library writes everywhere.
func (p Package) Var() string {
	if i := strings.LastIndex(p.Path, "/"); i >= 0 {
		return p.Path[i+1:]
	}
	return p.Path
}

// Find returns the entry at a dotted path inside the package, and whether it
// is there. "Mutex" is a top level name, "Mutex.Lock" a field of what Mutex
// builds, and so on for as many steps as the name has.
func (p Package) Find(path string) (Entry, bool) {
	if path == "" {
		return Entry{}, false
	}

	var (
		entries = p.Entries
		found   Entry
	)
	for _, name := range strings.Split(path, ".") {
		i := indexOf(entries, name)
		if i < 0 {
			return Entry{}, false
		}
		found = entries[i]
		entries = found.Children
	}
	return found, true
}

func indexOf(entries []Entry, name string) int {
	for i, e := range entries {
		if e.Name == name {
			return i
		}
	}
	return -1
}

// Load reads the module at path, which is a file or the directory of a module
// made of several. name is what the module is called, the thing an import
// asks for.
func Load(name, path string) (Package, error) {
	p := Package{Path: name}

	info, err := os.Stat(path)
	if err != nil {
		return p, err
	}

	if info.IsDir() {
		p.Dir = path
		if p.Files, err = mod.Files(path); err != nil {
			return p, err
		}
	} else {
		p.Files = []string{path}
	}
	if len(p.Files) == 0 {
		return p, fmt.Errorf("doc: %s holds no tau source", path)
	}

	var all []Entry

	for _, f := range p.Files {
		src, err := os.ReadFile(f)
		if err != nil {
			return p, err
		}

		toks := tokens(f, string(src))
		// The comment a file opens with belongs to the module. Of several
		// files, the first one to have one speaks for all of them, the way
		// one file of a Go package holds the package comment.
		if p.Doc == "" {
			p.Doc = header(toks)
		}
		all = append(all, entries(toks, 0, len(toks))...)
	}

	// Every top level name, the ones kept back as well: a constructor that is
	// not given away still says what the one that is hands out.
	index := make(map[string]Entry, len(all))
	for _, e := range all {
		index[e.Name] = e
	}

	p.Entries = exported(follow(all, index, nil))
	return p, nil
}

// follow gives an entry that builds nothing of its own the fields of whatever
// it hands back. The standard library is written that way over and over: Open
// returns what newFile built, Compile what makeRegexp did, and the name the
// reader has is the first of each pair.
//
// seen holds the names being followed, so that two functions returning each
// other's result stop rather than going round.
func follow(entries []Entry, index map[string]Entry, seen []string) []Entry {
	out := make([]Entry, len(entries))

	for i, e := range entries {
		e.Children = follow(e.Children, index, seen)

		// Nothing it gives away of its own: a local the body needed is not
		// something the reader of the documentation ever sees.
		if !givesAway(e.Children) {
			if children := handedBack(e.returns, index, seen); children != nil {
				e.Children = children
			}
		}
		out[i] = e
	}
	return out
}

// exported keeps the names a module gives away, at every depth.
func exported(entries []Entry) []Entry {
	var out []Entry

	for _, e := range entries {
		if !isExported(e.Name) {
			continue
		}
		e.Children = exported(e.Children)
		out = append(out, e)
	}
	return out
}

// givesAway reports whether any of these is a name a module hands out.
// handedBack walks the names an entry could be handing back until one of them
// builds something, and gives the fields of that one. It is a walk and not a
// single step because a name is handed on more than once: Compile returns what
// compile built, and compile what makeRegexp did, so the reader who writes
// Compile is three names away from the object they end up holding.
func handedBack(returns []string, index map[string]Entry, seen []string) []Entry {
	for _, name := range returns {
		if contains(seen, name) {
			continue
		}

		from, ok := index[name]
		if !ok {
			continue
		}

		seen = append(seen, name)
		if givesAway(from.Children) {
			return follow(from.Children, index, seen)
		}
		if children := handedBack(from.returns, index, seen); children != nil {
			return children
		}
	}
	return nil
}

func givesAway(entries []Entry) bool {
	for _, e := range entries {
		if isExported(e.Name) {
			return true
		}
	}
	return false
}

func contains(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

// Name is what to call the module found at path, when the caller has no
// better name for it: the directory of a module made of several files, the
// file without its extension otherwise.
func Name(path string) string {
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return filepath.Base(path)
	}
	return strings.TrimSuffix(filepath.Base(path), ".tau")
}

// tok is an item with the line it starts on, which is what tells a comment
// written above a name from one written beside it.
type tok struct {
	item.Item
	file  string
	line  int
	first bool // nothing but space before it on its line
}

func tokens(file, src string) []tok {
	var (
		out  []tok
		line = 1
		pos  int
		bol  = true // at the beginning of a line
	)

	for i := range lexer.Lex(src) {
		if i.Is(item.EOF) || i.Is(item.Error) {
			break
		}

		if i.Pos > pos && i.Pos <= len(src) {
			if n := strings.Count(src[pos:i.Pos], "\n"); n > 0 {
				line += n
				bol = true
			}
			pos = i.Pos
		}

		t := tok{Item: i, file: file, line: line, first: bol}
		if i.Is(item.RawString) {
			line += strings.Count(i.Val, "\n")
		}
		// A newline is not a token that sits on a line of its own: it is the
		// end of the one before it.
		if !t.isBreak() {
			bol = false
		}
		out = append(out, t)
	}

	return out
}

func (t tok) isBreak() bool {
	return t.Is(item.Semicolon) && t.Val != ";"
}

func (t tok) opens() bool {
	return t.Is(item.LBrace) || t.Is(item.LBracket) || t.Is(item.LParen)
}

func (t tok) closes() bool {
	return t.Is(item.RBrace) || t.Is(item.RBracket) || t.Is(item.RParen)
}

func (t tok) text() string {
	switch t.Typ {
	case item.String:
		return `"` + t.Val + `"`
	case item.RawString:
		return "`" + t.Val + "`"
	case item.Semicolon:
		return ";"
	default:
		return t.Val
	}
}

// header is the comment a file opens with, which is the one about the module
// rather than about any name in it. A comment followed straight away by a
// name is that name's, not the file's.
func header(toks []tok) string {
	var (
		block []tok
		last  int
	)

	for _, t := range toks {
		if t.Is(item.Comment) && t.first {
			// An empty line ends the block: what comes after it is about
			// whatever follows it, not about the file.
			if len(block) > 0 && t.line > last+1 {
				break
			}
			block = append(block, t)
			last = t.line
			continue
		}
		if t.isBreak() {
			continue
		}
		// The first thing that isn't a comment: the block was the file's only
		// if an empty line stands between them.
		if len(block) > 0 && t.line == last+1 {
			return ""
		}
		break
	}
	return comment(block)
}

// comment turns a run of comment tokens into the text they hold: the '#' and
// the single space after it taken off, the rest left as it was written, so
// that an indented example stays indented.
func comment(block []tok) string {
	if len(block) == 0 {
		return ""
	}

	lines := make([]string, len(block))
	for i, t := range block {
		s := strings.TrimPrefix(t.Val, "#")
		lines[i] = strings.TrimPrefix(s, " ")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), "\n")
}

// entries reads the names assigned between two token indexes, at the depth
// those indexes open on: the top level of a file, or the body of a function
// once it has been descended into.
func entries(toks []tok, start, end int) []Entry {
	var (
		out   []Entry
		block []tok
		last  int
		depth int
		i     = start
	)

	for i < end {
		t := toks[i]

		switch {
		case t.opens():
			depth++
			i++

		case t.closes():
			depth--
			i++

		case depth != 0:
			i++

		case t.isBreak():
			i++

		case t.Is(item.Comment):
			// Only a comment on a line of its own explains what comes next.
			// One written beside code is about that line.
			if t.first {
				if len(block) > 0 && t.line > last+1 {
					block = nil
				}
				block = append(block, t)
				last = t.line
			}
			i++

		default:
			e, next, ok := assignment(toks, i, end)
			if ok {
				// A comment separated from the name by an empty line was
				// about something else.
				if len(block) > 0 && e.Line == last+1 {
					e.Doc = comment(block)
				}
				out = append(out, e)
			}
			block = nil
			i = next
		}
	}

	return out
}

// assignment reads one statement. It returns an entry when the statement
// gives a name a value, and where the statement ends either way.
func assignment(toks []tok, i, end int) (Entry, int, bool) {
	var (
		e    Entry
		name = i
	)

	if !toks[i].Is(item.Ident) {
		return e, statement(toks, i, end), false
	}

	// A field of an object: "b.Write = ...". The name is the field, since
	// that is what the reader of the documentation will write after the dot.
	eq := i + 1
	if eq+1 < end && toks[eq].Is(item.Dot) && toks[eq+1].Is(item.Ident) {
		name = eq + 1
		eq = eq + 2
	}
	if eq >= end || !toks[eq].Is(item.Assign) {
		return e, statement(toks, i, end), false
	}

	e = Entry{
		Name: toks[name].Val,
		File: toks[name].file,
		Line: toks[name].line,
	}

	rhs := eq + 1
	stop := statement(toks, i, end)
	if rhs >= end {
		return e, stop, true
	}

	if toks[rhs].Is(item.Function) {
		e.Sig, e.Children, e.returns = function(toks, rhs, end)
		return e, stop, true
	}

	last := stop
	if last > end {
		last = end
	}
	e.Val = value(toks, rhs, last)
	// A name assigned the result of a call holds what that call builds.
	if toks[rhs].Is(item.Ident) && rhs+1 < end && toks[rhs+1].Is(item.LParen) {
		e.returns = []string{toks[rhs].Val}
	}
	return e, stop, true
}

// function reads a function literal: its parameters, written back as the
// signature, and the names its body gives, which are the fields of whatever
// object it builds.
func function(toks []tok, i, end int) (string, []Entry, []string) {
	// The parameter list.
	var (
		sig    strings.Builder
		params = i + 1
	)
	sig.WriteString("fn(")

	if params < end && toks[params].Is(item.LParen) {
		depth := 0
		j := params
		for ; j < end; j++ {
			if toks[j].opens() {
				depth++
			} else if toks[j].closes() {
				if depth--; depth == 0 {
					break
				}
			} else if depth == 1 && toks[j].Is(item.Comma) {
				sig.WriteString(", ")
			} else if depth == 1 {
				sig.WriteString(toks[j].text())
			}
		}
		i = j + 1
	}
	sig.WriteString(")")

	// The body, read by the same rule as the file around it.
	for j := i; j < end; j++ {
		if toks[j].Is(item.LBrace) {
			if close := match(toks, j, end); close > j {
				return sig.String(), entries(toks, j+1, close), returns(toks, j+1, close)
			}
			break
		}
		if !toks[j].isBreak() {
			break // a one line body with no braces holds no names
		}
	}
	return sig.String(), nil, nil
}

// returns are the names a body could be handing back: what a "return f(...)"
// calls, and for a "return x" the name of whatever x was last assigned from.
// Anything else is left out, since there is no name to follow.
func returns(toks []tok, start, end int) []string {
	var (
		out    []string
		locals = map[string]string{}
		depth  int
		// The last call the body makes with nothing written after it. A body
		// is worth the value of its last expression, so a function ending in
		// one hands back what it built, with no return to say so.
		last string
	)

	for i := start; i < end; i++ {
		switch {
		case toks[i].opens():
			depth++
			continue
		case toks[i].closes():
			depth--
			continue
		case depth != 0:
			continue
		}

		// "x = f(...)" remembers that x holds what f built.
		if toks[i].Is(item.Ident) && i+3 < end && toks[i+1].Is(item.Assign) &&
			toks[i+2].Is(item.Ident) && toks[i+3].Is(item.LParen) {
			locals[toks[i].Val] = toks[i+2].Val
			continue
		}

		if !toks[i].Is(item.Return) || i+1 >= end || !toks[i+1].Is(item.Ident) {
			// A call standing on its own, which is the value of the body when
			// nothing follows it.
			if toks[i].Is(item.Ident) && i+1 < end && toks[i+1].Is(item.LParen) {
				last = toks[i].Val
			}
			continue
		}

		name := toks[i+1].Val
		if i+2 < end && toks[i+2].Is(item.LParen) {
			out = append(out, name)
		} else if from, ok := locals[name]; ok {
			out = append(out, from)
		}
	}

	// A return says it outright and is believed first; the trailing call is
	// only what is left to go on when the body never says it.
	if len(out) == 0 && last != "" {
		out = append(out, last)
	}
	return out
}

// value writes back what a name that isn't a function was assigned, when it
// is short enough to read at a glance.
func value(toks []tok, i, end int) string {
	var out strings.Builder

	for j := i; j < end; j++ {
		if toks[j].Is(item.Comment) || toks[j].isBreak() {
			break
		}
		if out.Len() > 0 && wantsSpace(toks[j-1], toks[j]) {
			out.WriteByte(' ')
		}
		out.WriteString(toks[j].text())
		if out.Len() > 64 {
			return ""
		}
	}
	return out.String()
}

// wantsSpace is the spacing of the formatter, cut down to what writing back a
// short expression needs.
func wantsSpace(prev, cur tok) bool {
	switch {
	case prev.Is(item.LParen) || prev.Is(item.LBracket) || prev.Is(item.Dot):
		return false
	case cur.Is(item.RParen) || cur.Is(item.RBracket) || cur.Is(item.Comma) ||
		cur.Is(item.Dot) || cur.Is(item.LParen):
		return false
	case cur.Is(item.Colon):
		return false
	default:
		return true
	}
}

// statement is the index just past the end of the statement starting at i.
//
// What ends one is the next token to begin a line, as long as nothing it
// opened is still open: a line break is the end of a statement in tau, and a
// token at the start of a line is the proof that one happened. The lexer's
// own newline is not enough to go by, since a line ending in a comment gives
// none - the comment eats it - and a statement would then swallow everything
// down to the next line that has one.
func statement(toks []tok, i, end int) int {
	depth := 0

	for j := i; j < end; j++ {
		if j > i && depth <= 0 && toks[j].first {
			return j
		}

		switch {
		case toks[j].opens():
			depth++
		case toks[j].closes():
			depth--
		// A semicolon somebody wrote separates two statements on one line.
		case toks[j].Is(item.Semicolon) && !toks[j].isBreak() && depth <= 0:
			return j + 1
		}
	}
	return end
}

// match is the index of the bracket closing the one at i.
func match(toks []tok, i, end int) int {
	depth := 0

	for ; i < end; i++ {
		if toks[i].opens() {
			depth++
		} else if toks[i].closes() {
			if depth--; depth == 0 {
				return i
			}
		}
	}
	return -1
}

// isExported is the rule the interpreter itself goes by: a module gives away
// the names that start with an upper case letter.
func isExported(name string) bool {
	if name == "" {
		return false
	}
	return unicode.IsUpper([]rune(name)[0])
}
