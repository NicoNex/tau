package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/NicoNex/tau/internal/format"
	"github.com/NicoNex/tau/internal/parser"
	"github.com/NicoNex/tau/internal/tauerr"
)

type server struct {
	conn *conn
	docs map[string]*document

	initialized bool
	shutdown    bool
	exitCode    int
	done        bool
}

func newServer(c *conn) *server {
	return &server{conn: c, docs: make(map[string]*document)}
}

// run reads messages until the stream ends or `exit` arrives. Requests are
// served one at a time: an editor sends few of them, and a single goroutine
// costs nothing and cannot race over the buffers.
func (s *server) run() error {
	for !s.done {
		body, err := s.conn.read()
		if err != nil {
			return err
		}

		var req request
		if err := json.Unmarshal(body, &req); err != nil {
			s.conn.replyErr(nil, errParse, "invalid JSON: "+err.Error())
			continue
		}
		s.dispatch(req)
	}
	return nil
}

// dispatch routes one message. A handler that panics must not take the
// editor's language support down with it, so the panic becomes an error
// response and the loop carries on.
func (s *server) dispatch(req request) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("tau-lsp: panic in %s: %v", req.Method, r)
			if isRequest(req.ID) {
				s.conn.replyErr(req.ID, errInternal, fmt.Sprintf("internal error in %s", req.Method))
			}
		}
	}()

	if req.Method == "" {
		if isRequest(req.ID) {
			s.conn.replyErr(req.ID, errInvalidRequest, "request without a method")
		}
		return
	}

	// Everything but initialize/exit is refused until the handshake is done,
	// and everything but exit once shutdown has been asked for.
	switch {
	case !s.initialized && req.Method != "initialize" && req.Method != "exit":
		if isRequest(req.ID) {
			s.conn.replyErr(req.ID, errServerNotInit, "server not initialized")
		}
		return
	case s.shutdown && req.Method != "exit":
		if isRequest(req.ID) {
			s.conn.replyErr(req.ID, errInvalidRequest, "server is shutting down")
		}
		return
	}

	switch req.Method {
	case "initialize":
		s.initialized = true
		s.conn.reply(req.ID, s.capabilities())

	case "initialized":
		// Nothing to do, the client is only saying it is ready.

	case "shutdown":
		s.shutdown = true
		s.conn.reply(req.ID, nil)

	case "exit":
		if !s.shutdown {
			s.exitCode = 1
		}
		s.done = true

	case "textDocument/didOpen":
		s.didOpen(req.Params)
	case "textDocument/didChange":
		s.didChange(req.Params)
	case "textDocument/didClose":
		s.didClose(req.Params)
	case "textDocument/didSave":
		// The buffer is already up to date from didChange.

	case "textDocument/completion":
		s.conn.reply(req.ID, s.completion(req.Params))
	case "textDocument/hover":
		s.conn.reply(req.ID, s.hover(req.Params))
	case "textDocument/definition":
		s.conn.reply(req.ID, s.definition(req.Params))
	case "textDocument/documentSymbol":
		s.conn.reply(req.ID, s.documentSymbol(req.Params))
	case "textDocument/formatting":
		result, err := s.formatting(req.Params)
		if err != nil {
			// A file that does not parse is not an error the user needs a
			// popup for: the diagnostics already say what is wrong.
			log.Printf("tau-lsp: formatting: %v", err)
			s.conn.reply(req.ID, nil)
			return
		}
		s.conn.reply(req.ID, result)

	default:
		if isRequest(req.ID) {
			s.conn.replyErr(req.ID, errMethodNotFound, "method not found: "+req.Method)
		}
	}
}

// capabilities advertises what this server actually answers, and nothing
// else: a capability claimed but not implemented shows up to the user as an
// editor feature that silently does nothing.
func (s *server) capabilities() map[string]any {
	return map[string]any{
		"capabilities": map[string]any{
			"textDocumentSync": map[string]any{
				"openClose": true,
				"change":    1, // full text on every change
			},
			"completionProvider": map[string]any{
				"triggerCharacters": []string{"."},
			},
			"hoverProvider":              true,
			"definitionProvider":         true,
			"documentSymbolProvider":     true,
			"documentFormattingProvider": true,
		},
		"serverInfo": map[string]any{"name": "tau-lsp", "version": "1"},
	}
}

/* =========================
   Document synchronisation
   ========================= */

type textDocumentIdentifier struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
}

type textDocumentPositionParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	Position     position               `json:"position"`
}

func (s *server) didOpen(params json.RawMessage) {
	var p struct {
		TextDocument struct {
			URI        string `json:"uri"`
			LanguageID string `json:"languageId"`
			Version    int    `json:"version"`
			Text       string `json:"text"`
		} `json:"textDocument"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		log.Printf("tau-lsp: didOpen: %v", err)
		return
	}

	doc := newDocument(p.TextDocument.URI, p.TextDocument.Text, p.TextDocument.Version)
	s.docs[doc.uri] = doc
	s.publishDiagnostics(doc)
}

func (s *server) didChange(params json.RawMessage) {
	var p struct {
		TextDocument   textDocumentIdentifier `json:"textDocument"`
		ContentChanges []struct {
			Text string `json:"text"`
		} `json:"contentChanges"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		log.Printf("tau-lsp: didChange: %v", err)
		return
	}
	if len(p.ContentChanges) == 0 {
		return
	}

	doc := s.docs[p.TextDocument.URI]
	if doc == nil {
		doc = newDocument(p.TextDocument.URI, "", p.TextDocument.Version)
		s.docs[doc.uri] = doc
	}
	// Full sync was advertised, so the last change carries the whole buffer.
	doc.version = p.TextDocument.Version
	doc.setText(p.ContentChanges[len(p.ContentChanges)-1].Text)
	s.publishDiagnostics(doc)
}

func (s *server) didClose(params json.RawMessage) {
	var p struct {
		TextDocument textDocumentIdentifier `json:"textDocument"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return
	}
	delete(s.docs, p.TextDocument.URI)
	// The client keeps showing diagnostics until they are cleared.
	s.conn.notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         p.TextDocument.URI,
		"diagnostics": []any{},
	})
}

func (s *server) document(uri string) *document {
	return s.docs[uri]
}

/* =========================
   Diagnostics
   ========================= */

type diagnostic struct {
	Range    textRange `json:"range"`
	Severity int       `json:"severity"`
	Source   string    `json:"source"`
	Message  string    `json:"message"`
}

func (s *server) publishDiagnostics(doc *document) {
	s.conn.notify("textDocument/publishDiagnostics", map[string]any{
		"uri":         doc.uri,
		"version":     doc.version,
		"diagnostics": diagnose(doc),
	})
}

// diagnose runs the real parser over the buffer and turns each error into a
// range the editor can underline.
func diagnose(doc *document) []diagnostic {
	name := doc.path
	if name == "" {
		name = doc.uri
	}

	_, errs := parser.Parse(name, doc.text)
	out := make([]diagnostic, 0, len(errs))

	seen := make(map[textRange]bool)
	for _, err := range errs {
		r, msg := errorRange(doc, err)
		if seen[r] {
			continue
		}
		seen[r] = true
		out = append(out, diagnostic{Range: r, Severity: 1, Source: "tau", Message: msg})
	}
	return out
}

// errorRange places a parse error in the buffer. tauerr reports a one based
// line and a column counted from the first non blank character of that line,
// because that is how it draws the caret under the source it prints back;
// putting it back where it belongs means adding the indentation again.
func errorRange(doc *document, err error) (textRange, string) {
	te, ok := err.(tauerr.TauErr)
	if !ok {
		return textRange{}, err.Error()
	}

	line := te.Line - 1
	if line < 0 {
		line = 0
	}
	if line >= len(doc.lines) {
		line = len(doc.lines) - 1
	}

	start, end := doc.lineRange(line)
	text := doc.text[start:end]
	indent := len(text) - len(strings.TrimLeft(text, " \t"))

	from := start + indent + te.Column
	if from > end {
		from = end
	}
	if from < start {
		from = start
	}

	// Underline the token the parser stopped on, or the rest of the line
	// when there is none, so that the squiggle has a width.
	to := from
	for to < end && isWordByte(doc.text[to]) {
		to++
	}
	if to == from {
		if from < end {
			to = from + 1
		} else if from > start {
			from = start + indent
		}
	}

	return doc.rangeOf(from, to), message(te)
}

// message is the last line of a TauErr, the one saying what went wrong. The
// rest is the source and the caret, which the editor draws itself.
func message(te tauerr.TauErr) string {
	msg := te.Message
	if i := strings.LastIndex(msg, "^\n"); i >= 0 {
		msg = msg[i+2:]
	}
	return strings.TrimSpace(msg)
}

func isWordByte(b byte) bool {
	return b == '_' || b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
}

/* =========================
   Formatting
   ========================= */

func (s *server) formatting(params json.RawMessage) (any, error) {
	var p struct {
		TextDocument textDocumentIdentifier `json:"textDocument"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}
	doc := s.document(p.TextDocument.URI)
	if doc == nil {
		return nil, fmt.Errorf("no such document: %s", p.TextDocument.URI)
	}

	name := doc.path
	if name == "" {
		name = doc.uri
	}
	out, err := format.Source(name, doc.text)
	if err != nil {
		return nil, err
	}
	if out == doc.text {
		return []any{}, nil
	}

	// One edit replacing everything: the formatter works on whole files and
	// a diff would only be a guess at what it did.
	return []map[string]any{{
		"range":   doc.rangeOf(0, len(doc.text)),
		"newText": out,
	}}, nil
}
