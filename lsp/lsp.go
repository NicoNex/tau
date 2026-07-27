package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"unicode/utf8"
)

/* =========================
   JSON-RPC envelope types
   ========================= */

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *respError      `json:"error,omitempty"`
}

type respError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

/* =========================
   LSP minimal structs
   ========================= */

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}
type rng struct {
	Start position `json:"start"`
	End   position `json:"end"`
}
type diagnostic struct {
	Range    rng    `json:"range"`
	Severity int    `json:"severity,omitempty"` // 1=Error,2=Warning
	Source   string `json:"source,omitempty"`
	Message  string `json:"message"`
}

/* =========================
   Server & docs
   ========================= */

type document struct {
	URI        string
	Version    int
	Text       string
	lineStarts []int // byte offsets of each line start
}

func (d *document) rebuildLineIndex() {
	d.lineStarts = d.lineStarts[:0]
	d.lineStarts = append(d.lineStarts, 0)
	for i := 0; i < len(d.Text); i++ {
		if d.Text[i] == '\n' {
			d.lineStarts = append(d.lineStarts, i+1)
		}
	}
}

func (d *document) offsetAt(p position) int {
	if p.Line < 0 {
		return 0
	}
	if p.Line >= len(d.lineStarts) {
		return len(d.Text)
	}
	base := d.lineStarts[p.Line]
	// clamp within line
	i := base
	end := len(d.Text)
	if p.Line+1 < len(d.lineStarts) {
		end = d.lineStarts[p.Line+1]
	}
	target := base + p.Character
	if target < i {
		return i
	}
	if target > end {
		return end
	}
	return target
}

type server struct {
	in  *bufio.Reader
	out *bufio.Writer

	mu   sync.RWMutex
	docs map[string]*document
}

func newServer() *server {
	return &server{
		in:   bufio.NewReader(os.Stdin),
		out:  bufio.NewWriter(os.Stdout),
		docs: make(map[string]*document),
	}
}

/* =========================
   Transport (fast stdio)
   ========================= */

func (s *server) readMessage() ([]byte, error) {
	var contentLength int
	// Read headers (LSP: CRLF lines, blank line terminates)
	for {
		line, err := s.in.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		// case-insensitive
		if len(line) >= 15 && strings.EqualFold(line[:15], "content-length:") {
			fmt.Sscanf(line, "Content-Length: %d", &contentLength)
		}
	}
	if contentLength <= 0 {
		return nil, io.EOF
	}
	buf := make([]byte, contentLength)
	if _, err := io.ReadFull(s.in, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (s *server) sendJSON(v any) {
	data, _ := json.Marshal(v)
	fmt.Fprintf(s.out, "Content-Length: %d\r\n\r\n", len(data))
	_, _ = s.out.Write(data)
	_ = s.out.Flush()
}

func (s *server) respond(id json.RawMessage, result any) {
	s.sendJSON(response{JSONRPC: "2.0", ID: id, Result: result})
}
func (s *server) respondErr(id json.RawMessage, code int, msg string) {
	s.sendJSON(response{JSONRPC: "2.0", ID: id, Error: &respError{Code: code, Message: msg}})
}
func (s *server) notify(method string, params any) {
	env := map[string]any{"jsonrpc": "2.0", "method": method, "params": params}
	s.sendJSON(env)
}

/* =========================
   Tau scanner / diagnostics
   ========================= */

type tauDiag struct {
	msg       string
	line, col int // 0-based
}

type brace struct {
	ch        rune
	line, col int
}

type scanner struct {
	src       string
	i         int
	line, col int
	inComment bool
	inQuotes  bool
	inRaw     bool
	esc       bool
	stack     []brace
	diags     []tauDiag
}

var diagPool = sync.Pool{New: func() any { return make([]tauDiag, 0, 16) }}
var stackPool = sync.Pool{New: func() any { return make([]brace, 0, 16) }}

func lintTauFast(src string) []tauDiag {
	sc := scanner{
		src:   src,
		diags: diagPool.Get().([]tauDiag)[:0],
		stack: stackPool.Get().([]brace)[:0],
	}
	defer func() {
		diagPool.Put(sc.diags[:0])
		stackPool.Put(sc.stack[:0])
	}()

	for sc.i < len(sc.src) {
		r, w := utf8.DecodeRuneInString(sc.src[sc.i:])
		if r == utf8.RuneError && w == 1 {
			r = rune(sc.src[sc.i]) // treat as byte
		}
		sc.advance(r, w)

		// inside comment
		if sc.inComment {
			if r == '\n' {
				sc.inComment = false
			}
			continue
		}
		// raw string
		if sc.inRaw {
			if r == '`' {
				sc.inRaw = false
			}
			continue
		}
		// quoted string
		if sc.inQuotes {
			if sc.esc {
				sc.esc = false
				continue
			}
			if r == '\\' {
				sc.esc = true
				continue
			}
			if r == '"' {
				sc.inQuotes = false
			}
			continue
		}

		// outside string/comment
		switch r {
		case '#':
			sc.inComment = true
		case '"':
			sc.inQuotes = true
		case '`':
			sc.inRaw = true
		case '{', '(', '[':
			sc.stack = append(sc.stack, brace{ch: r, line: sc.line, col: sc.col})
		case '}', ')', ']':
			sc.matchClose(r)
		}
	}

	// EOF checks
	if sc.inQuotes {
		sc.diags = append(sc.diags, tauDiag{"unterminated string literal", sc.line, sc.col})
	}
	if sc.inRaw {
		sc.diags = append(sc.diags, tauDiag{"unterminated raw string literal", sc.line, sc.col})
	}
	for _, b := range sc.stack {
		sc.diags = append(sc.diags, tauDiag{fmt.Sprintf("unclosed '%c'", b.ch), b.line, b.col})
	}

	// copy out (avoid returning pooled slice)
	out := make([]tauDiag, len(sc.diags))
	copy(out, sc.diags)
	return out
}

func (sc *scanner) advance(r rune, w int) {
	sc.i += w
	if r == '\n' {
		sc.line++
		sc.col = 0
	} else {
		sc.col++
	}
}

func (sc *scanner) matchClose(r rune) {
	if len(sc.stack) == 0 {
		sc.diags = append(sc.diags, tauDiag{fmt.Sprintf("stray closing '%c'", r), sc.line, sc.col})
		return
	}
	top := sc.stack[len(sc.stack)-1]
	match := (top.ch == '{' && r == '}') || (top.ch == '(' && r == ')') || (top.ch == '[' && r == ']')
	if !match {
		sc.diags = append(sc.diags, tauDiag{
			fmt.Sprintf("mismatched closing '%c' (opened with '%c')", r, top.ch),
			sc.line, sc.col,
		})
		return
	}
	sc.stack = sc.stack[:len(sc.stack)-1]
}

/* =========================
   Hover / Completion utils
   ========================= */

var keywords = []string{
	"fn", "return", "if", "else", "for", "true", "false", "null",
}

var builtins = []string{
	"len", "println", "print", "input", "string", "error", "type",
	"int", "float", "exit", "append", "new", "failed", "plugin",
	"hex", "oct", "bin", "slice", "keys", "delete", "bytes", "import",
	"pipe", "send", "recv", "close",
}

var (
	kwSet = toSet(keywords)
	biSet = toSet(builtins)
)

func toSet(xs []string) map[string]struct{} {
	m := make(map[string]struct{}, len(xs))
	for _, x := range xs {
		m[x] = struct{}{}
	}
	return m
}

func wordAt(doc *document, pos position) string {
	off := doc.offsetAt(pos)
	// find line bounds
	lineIdx := 0
	for i := 0; i+1 < len(doc.lineStarts); i++ {
		if doc.lineStarts[i] <= off && off < doc.lineStarts[i+1] {
			lineIdx = i
			break
		}
		if i+1 == len(doc.lineStarts)-1 && off >= doc.lineStarts[i+1] {
			lineIdx = i + 1
		}
	}
	start := doc.lineStarts[lineIdx]
	end := len(doc.Text)
	if lineIdx+1 < len(doc.lineStarts) {
		end = doc.lineStarts[lineIdx+1]
	}
	line := doc.Text[start:end]

	rel := off - start
	if rel < 0 || rel > len(line) {
		return ""
	}
	isWord := func(b byte) bool {
		return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
	}
	l, r := rel, rel
	for l > 0 && isWord(line[l-1]) {
		l--
	}
	for r < len(line) && isWord(line[r]) {
		r++
	}
	return line[l:r]
}

/* =========================
   Handlers
   ========================= */

func (s *server) handleInitialize(id json.RawMessage) {
	s.respond(id, map[string]any{
		"capabilities": map[string]any{
			"textDocumentSync": map[string]any{"openClose": true, "change": 2},
			"hoverProvider":    true,
			"completionProvider": map[string]any{
				"triggerCharacters": []string{".", "\"", "#", "{", "["},
			},
		},
	})
}

func (s *server) handleDidOpen(params json.RawMessage) {
	var p struct {
		TextDocument struct {
			URI      string `json:"uri"`
			Version  int    `json:"version"`
			Text     string `json:"text"`
			Language string `json:"languageId"`
		} `json:"textDocument"`
	}
	_ = json.Unmarshal(params, &p)

	doc := &document{URI: p.TextDocument.URI, Version: p.TextDocument.Version, Text: p.TextDocument.Text}
	doc.rebuildLineIndex()

	s.mu.Lock()
	s.docs[doc.URI] = doc
	s.mu.Unlock()

	s.publishDiagnostics(doc)
}

func (s *server) handleDidChange(params json.RawMessage) {
	var p struct {
		TextDocument struct {
			URI     string `json:"uri"`
			Version int    `json:"version"`
		} `json:"textDocument"`
		ContentChanges []struct {
			Text string `json:"text"`
		} `json:"contentChanges"`
	}
	_ = json.Unmarshal(params, &p)

	s.mu.Lock()
	doc := s.docs[p.TextDocument.URI]
	if doc == nil {
		doc = &document{URI: p.TextDocument.URI}
		s.docs[p.TextDocument.URI] = doc
	}
	doc.Version = p.TextDocument.Version
	if n := len(p.ContentChanges); n > 0 {
		doc.Text = p.ContentChanges[n-1].Text // treat as full sync
		doc.rebuildLineIndex()
	}
	s.mu.Unlock()

	s.publishDiagnostics(doc)
}

func (s *server) handleDidClose(params json.RawMessage) {
	var p struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	_ = json.Unmarshal(params, &p)
	s.mu.Lock()
	delete(s.docs, p.TextDocument.URI)
	s.mu.Unlock()
	// clear diags
	s.notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         p.TextDocument.URI,
		"version":     0,
		"diagnostics": []diagnostic{},
	})
}

func (s *server) handleHover(id json.RawMessage, params json.RawMessage) {
	var p struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
		Position position `json:"position"`
	}
	_ = json.Unmarshal(params, &p)

	s.mu.RLock()
	doc := s.docs[p.TextDocument.URI]
	s.mu.RUnlock()

	var contents any = nil
	if doc != nil {
		w := wordAt(doc, p.Position)
		if w != "" {
			if _, ok := kwSet[w]; ok {
				contents = map[string]any{"kind": "markdown", "value": fmt.Sprintf("**keyword** `%s`", w)}
			} else if _, ok := biSet[w]; ok {
				contents = map[string]any{"kind": "markdown", "value": fmt.Sprintf("**builtin** `%s(...)`", w)}
			}
		}
	}
	s.respond(id, map[string]any{"contents": contents})
}

func (s *server) handleCompletion(id json.RawMessage) {
	items := make([]map[string]any, 0, len(keywords)+len(builtins)+3)
	for _, k := range keywords {
		items = append(items, map[string]any{"label": k, "kind": 14, "insertText": k})
	}
	for _, b := range builtins {
		items = append(items, map[string]any{"label": b, "kind": 3, "insertText": b + "()"})
	}
	items = append(items,
		map[string]any{"label": "fn", "kind": 15, "insertTextFormat": 2, "insertText": "fn(${1:name}) {\n\t$0\n}"},
		map[string]any{"label": "if", "kind": 15, "insertTextFormat": 2, "insertText": "if ${1:cond} {\n\t$0\n}"},
		map[string]any{"label": "if-else", "kind": 15, "insertTextFormat": 2, "insertText": "if ${1:cond} {\n\t$2\n} else {\n\t$0\n}"},
	)
	s.respond(id, map[string]any{"isIncomplete": false, "items": items})
}

func (s *server) publishDiagnostics(doc *document) {
	tds := lintTauFast(doc.Text)
	out := make([]diagnostic, 0, len(tds))
	for _, d := range tds {
		p := position{Line: d.line, Character: d.col}
		out = append(out, diagnostic{
			Range:    rng{Start: p, End: p},
			Severity: 1,
			Source:   "tau-lsp",
			Message:  d.msg,
		})
	}
	s.notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         doc.URI,
		"version":     doc.Version,
		"diagnostics": out,
	})
}

/* =========================
   Main loop
   ========================= */

func (s *server) route(req request) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req.ID)
	case "initialized":
		// no-op
	case "shutdown":
		s.respond(req.ID, nil)
	case "exit":
		os.Exit(0)
	case "textDocument/didOpen":
		s.handleDidOpen(req.Params)
	case "textDocument/didChange":
		s.handleDidChange(req.Params)
	case "textDocument/didClose":
		s.handleDidClose(req.Params)
	case "textDocument/hover":
		s.handleHover(req.ID, req.Params)
	case "textDocument/completion":
		s.handleCompletion(req.ID)
	default:
		if len(bytes.TrimSpace(req.ID)) > 0 {
			s.respondErr(req.ID, -32601, "method not found: "+req.Method)
		}
	}
}

func main() {
	s := newServer()
	for {
		raw, err := s.readMessage()
		if err != nil {
			if err == io.EOF {
				return
			}
			continue
		}
		var req request
		if err := json.Unmarshal(raw, &req); err != nil {
			// ignore malformed frames
			continue
		}
		s.route(req)
	}
}
