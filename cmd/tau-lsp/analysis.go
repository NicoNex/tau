package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/NicoNex/tau/internal/item"
	"github.com/NicoNex/tau/internal/lexer"
)

// symbolKind mirrors the handful of LSP SymbolKind values a tau file can
// produce. Everything is a name bound to a value, the kind only says what
// kind of value it was bound to.
type symbolKind int

const (
	kindVariable symbolKind = 13
	kindFunction symbolKind = 12
	kindModule   symbolKind = 2
)

// symbol is a name defined by a top level assignment.
type symbol struct {
	name   string
	kind   symbolKind
	pos    int    // byte offset of the name
	end    int    // byte offset just past the name
	detail string // parameter list for a function, module path for an import
	doc    string // the '#' comment block written above it
	// module, for an import, is the resolved path of the imported file.
	module string
}

// fileInfo is everything the server knows about one buffer, worked out from
// the token stream. The tokens are used rather than the AST because the AST
// keeps neither the comments nor the identifiers' positions in a form a tool
// can read back.
type fileInfo struct {
	symbols []symbol
	// byName is the last definition of each name, which is the one a jump
	// should land on when a name is assigned more than once.
	byName map[string]*symbol
	// imports maps the local name of a module to its resolved path.
	imports map[string]string
}

// analyse reads the token stream of src and picks out the top level
// definitions, their doc comments and the modules the file imports.
func analyse(path, src string) *fileInfo {
	info := &fileInfo{
		byName:  make(map[string]*symbol),
		imports: make(map[string]string),
	}

	toks := lex(src)
	comments := commentBlocks(src, toks)

	var depth int
	for i, t := range toks {
		switch t.Typ {
		case item.LBrace, item.LParen, item.LBracket:
			depth++
			continue
		case item.RBrace, item.RParen, item.RBracket:
			if depth > 0 {
				depth--
			}
			continue
		}

		// A definition is `name =` written at the outermost level. `==` and
		// the compound assignments are different tokens, so a plain Assign
		// after an identifier can only be a binding.
		if depth != 0 || !t.Is(item.Ident) || i+1 >= len(toks) || !toks[i+1].Is(item.Assign) {
			continue
		}
		// `a.b = x` and `a[i] = x` set a field, they do not define a name.
		if i > 0 && (toks[i-1].Is(item.Dot) || toks[i-1].Is(item.RBracket)) {
			continue
		}

		s := symbol{
			name: t.Val,
			kind: kindVariable,
			pos:  t.Pos,
			end:  t.Pos + len(t.Val),
			doc:  comments[lineOf(src, t.Pos)],
		}

		switch rest := toks[i+2:]; {
		case len(rest) > 0 && rest[0].Is(item.Function):
			s.kind = kindFunction
			s.detail = params(rest)

		case len(rest) > 1 && rest[0].Is(item.Import) && rest[1].Is(item.LParen):
			s.kind = kindModule
			if len(rest) > 2 && rest[2].Is(item.String) {
				name := unquote(rest[2].Val)
				s.detail = name
				s.module = resolveModule(path, name)
			}
		}

		info.symbols = append(info.symbols, s)
	}

	for i := range info.symbols {
		s := &info.symbols[i]
		info.byName[s.name] = s
		if s.kind == kindModule && s.module != "" {
			info.imports[s.name] = s.module
		}
	}
	return info
}

// lex drains the whole token stream. The channel must be read to the end or
// the lexer's goroutine is left blocked on a send forever.
func lex(src string) []item.Item {
	var out []item.Item
	for t := range lexer.Lex(src) {
		if t.Is(item.EOF) {
			break
		}
		out = append(out, t)
	}
	return out
}

// commentBlocks maps the line of each statement to the run of comment lines
// written immediately above it, which is where a tau file keeps its docs.
func commentBlocks(src string, toks []item.Item) map[int]string {
	// line -> comment text, for the comment lines that stand on their own.
	byLine := make(map[int]string)
	for _, t := range toks {
		if !t.Is(item.Comment) {
			continue
		}
		line := lineOf(src, t.Pos)
		// A comment sharing a line with code documents nothing.
		if strings.TrimSpace(src[lineStart(src, t.Pos):t.Pos]) != "" {
			continue
		}
		byLine[line] = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(t.Val), "#"))
	}

	docs := make(map[int]string, len(byLine))
	for line := range byLine {
		// The block belongs to the first line below it that is not a comment.
		if _, ok := byLine[line+1]; ok {
			continue
		}
		var block []string
		for l := line; ; l-- {
			c, ok := byLine[l]
			if !ok {
				break
			}
			block = append([]string{c}, block...)
		}
		docs[line+1] = strings.TrimSpace(strings.Join(block, "\n"))
	}
	return docs
}

// params renders the parameter list of a function literal starting at
// toks[0], for the one line signature shown in hover and completion.
func params(toks []item.Item) string {
	if len(toks) < 2 || !toks[1].Is(item.LParen) {
		return "fn()"
	}
	var names []string
	for _, t := range toks[2:] {
		if t.Is(item.RParen) {
			break
		}
		if t.Is(item.Ident) {
			names = append(names, t.Val)
		}
	}
	return "fn(" + strings.Join(names, ", ") + ")"
}

func unquote(s string) string {
	if v, err := strconv.Unquote(s); err == nil {
		return v
	}
	return strings.Trim(s, `"`)
}

func lineStart(src string, pos int) int {
	if i := strings.LastIndexByte(src[:min(pos, len(src))], '\n'); i >= 0 {
		return i + 1
	}
	return 0
}

// lineOf is the zero based line holding pos.
func lineOf(src string, pos int) int {
	return strings.Count(src[:min(pos, len(src))], "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}

/* =========================
   Module resolution
   ========================= */

// searchDirs are the directories a module is looked up into, in the order
// the runtime looks into them. This repeats the few lines of
// internal/vm.searchDirs on purpose: importing the vm would drag the whole
// cgo runtime into a program that only ever reads source.
func searchDirs(dir string) []string {
	var dirs []string

	if taupath := os.Getenv("TAUPATH"); taupath != "" {
		dirs = append(dirs, filepath.SplitList(taupath)...)
	}
	dirs = append(dirs, dir)

	if home, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(home, ".local", "lib", "tau"))
	}
	return append(dirs, "/usr/local/lib/tau", "/lib/tau")
}

// resolveModule finds the file that `import(name)` written in importer would
// load. Only sources are of interest: a compiled module holds no names a
// tool could read.
func resolveModule(importer, name string) string {
	if name == "" {
		return ""
	}
	dir := filepath.Dir(importer)

	candidates := func(base string) []string {
		if filepath.Ext(base) != "" {
			return []string{base}
		}
		return []string{base + ".tau"}
	}

	if filepath.IsAbs(name) {
		for _, p := range candidates(filepath.Clean(name)) {
			if isFile(p) {
				return p
			}
		}
		return ""
	}

	for _, p := range candidates(filepath.Clean(name)) {
		if isFile(p) {
			if abs, err := filepath.Abs(p); err == nil {
				return abs
			}
			return p
		}
	}
	for _, d := range searchDirs(dir) {
		for _, p := range candidates(filepath.Join(d, name)) {
			if isFile(p) {
				return p
			}
		}
	}
	return ""
}

func isFile(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// moduleCache keeps the analysis of the imported files, keyed by path and
// invalidated when the file changes on disk. Completion asks for a module's
// members on every keystroke and reparsing the standard library each time
// would be felt.
var moduleCache = struct {
	sync.Mutex
	entries map[string]*moduleEntry
}{entries: make(map[string]*moduleEntry)}

type moduleEntry struct {
	modTime int64
	size    int64
	info    *fileInfo
	src     string
}

// moduleInfo analyses the module at path, from the cache when it can.
func moduleInfo(path string) (*fileInfo, string) {
	if path == "" {
		return nil, ""
	}
	fi, err := os.Stat(path)
	if err != nil {
		return nil, ""
	}

	moduleCache.Lock()
	defer moduleCache.Unlock()

	if e, ok := moduleCache.entries[path]; ok && e.modTime == fi.ModTime().UnixNano() && e.size == fi.Size() {
		return e.info, e.src
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, ""
	}
	src := string(b)
	e := &moduleEntry{
		modTime: fi.ModTime().UnixNano(),
		size:    fi.Size(),
		info:    analyse(path, src),
		src:     src,
	}
	moduleCache.entries[path] = e
	return e.info, e.src
}

// exported are the members of a module a file importing it can name.
func exported(info *fileInfo) []symbol {
	if info == nil {
		return nil
	}
	var out []symbol
	for _, s := range info.symbols {
		if isExported(s.name) {
			out = append(out, s)
		}
	}
	return out
}
