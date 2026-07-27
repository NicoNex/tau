package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// client drives the server the way an editor does: framed JSON-RPC written
// to its input, framed JSON-RPC read back from its output. Nothing here
// knows about the server's internals, so a test that passes is a test an
// editor would also have passed.
type client struct {
	t   *testing.T
	in  bytes.Buffer
	out bytes.Buffer
	id  int
}

func newClient(t *testing.T) *client {
	return &client{t: t}
}

func (c *client) request(method string, params any) int {
	c.id++
	c.send(map[string]any{"jsonrpc": "2.0", "id": c.id, "method": method, "params": params})
	return c.id
}

func (c *client) notify(method string, params any) {
	c.send(map[string]any{"jsonrpc": "2.0", "method": method, "params": params})
}

func (c *client) send(msg any) {
	c.t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		c.t.Fatal(err)
	}
	fmt.Fprintf(&c.in, "Content-Length: %d\r\n\r\n%s", len(data), data)
}

// run feeds everything written so far to a fresh server and returns the
// messages it wrote back.
func (c *client) run() []map[string]any {
	c.t.Helper()
	c.notify("exit", nil)

	s := newServer(&conn{
		in:  bufio.NewReader(bytes.NewReader(c.in.Bytes())),
		out: bufio.NewWriter(&c.out),
	})
	if err := s.run(); err != nil && err != io.EOF {
		c.t.Fatalf("server: %v", err)
	}
	return decodeAll(c.t, c.out.Bytes())
}

func decodeAll(t *testing.T, raw []byte) []map[string]any {
	t.Helper()
	var out []map[string]any
	r := bufio.NewReader(bytes.NewReader(raw))

	for {
		length := -1
		for {
			line, err := r.ReadString('\n')
			if err == io.EOF && strings.TrimSpace(line) == "" {
				return out
			}
			if err != nil && err != io.EOF {
				t.Fatal(err)
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				break
			}
			if name, value, ok := strings.Cut(line, ":"); ok && strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
				n, err := strconv.Atoi(strings.TrimSpace(value))
				if err != nil {
					t.Fatalf("bad Content-Length %q", value)
				}
				length = n
			}
		}
		if length < 0 {
			return out
		}
		body := make([]byte, length)
		if _, err := io.ReadFull(r, body); err != nil {
			t.Fatalf("short body: %v", err)
		}
		var m map[string]any
		if err := json.Unmarshal(body, &m); err != nil {
			t.Fatalf("bad JSON from server: %v: %s", err, body)
		}
		out = append(out, m)
	}
}

func findResponse(t *testing.T, msgs []map[string]any, id int) map[string]any {
	t.Helper()
	for _, m := range msgs {
		if v, ok := m["id"]; ok {
			if n, ok := v.(float64); ok && int(n) == id {
				return m
			}
		}
	}
	t.Fatalf("no response with id %d in %v", id, msgs)
	return nil
}

func findNotification(msgs []map[string]any, method string) map[string]any {
	for _, m := range msgs {
		if m["method"] == method {
			return m
		}
	}
	return nil
}

func (c *client) open(uri, text string) {
	c.notify("textDocument/didOpen", map[string]any{
		"textDocument": map[string]any{
			"uri": uri, "languageId": "tau", "version": 1, "text": text,
		},
	})
}

func at(line, char int) map[string]any {
	return map[string]any{"line": line, "character": char}
}

func doc(uri string) map[string]any {
	return map[string]any{"uri": uri}
}

const testURI = "file:///tmp/lsp_test_main.tau"

func initialize(c *client) {
	c.request("initialize", map[string]any{"processId": nil, "rootUri": nil})
	c.notify("initialized", map[string]any{})
}

func TestInitializeAnswers(t *testing.T) {
	c := newClient(t)
	id := c.request("initialize", map[string]any{"processId": nil})
	msgs := c.run()

	res, ok := findResponse(t, msgs, id)["result"].(map[string]any)
	if !ok {
		t.Fatalf("initialize returned no result")
	}
	caps, ok := res["capabilities"].(map[string]any)
	if !ok {
		t.Fatalf("initialize returned no capabilities: %v", res)
	}
	for _, want := range []string{
		"textDocumentSync", "completionProvider", "hoverProvider",
		"definitionProvider", "documentSymbolProvider", "documentFormattingProvider",
	} {
		if _, ok := caps[want]; !ok {
			t.Errorf("capability %s not advertised", want)
		}
	}
}

func TestRequestsBeforeInitializeAreRefused(t *testing.T) {
	c := newClient(t)
	id := c.request("textDocument/hover", map[string]any{})
	msgs := c.run()

	if _, ok := findResponse(t, msgs, id)["error"]; !ok {
		t.Fatalf("a request before initialize should be an error")
	}
}

func TestMalformedInputDoesNotKillTheServer(t *testing.T) {
	c := newClient(t)
	initialize(c)
	// A frame that is not JSON at all, then a request that must still work.
	body := "{not json"
	fmt.Fprintf(&c.in, "Content-Length: %d\r\n\r\n%s", len(body), body)
	c.request("textDocument/documentSymbol", map[string]any{"textDocument": doc("file:///nope.tau")})
	id := c.request("initialize", map[string]any{})

	msgs := c.run()
	if findResponse(t, msgs, id) == nil {
		t.Fatal("server stopped answering after a malformed message")
	}
}

func TestUnknownMethodIsAnError(t *testing.T) {
	c := newClient(t)
	initialize(c)
	id := c.request("textDocument/nonsense", map[string]any{})
	msgs := c.run()

	e, ok := findResponse(t, msgs, id)["error"].(map[string]any)
	if !ok {
		t.Fatal("unknown method should return an error")
	}
	if int(e["code"].(float64)) != errMethodNotFound {
		t.Errorf("code = %v, want %d", e["code"], errMethodNotFound)
	}
}

func TestDiagnosticsPointAtTheBrokenLine(t *testing.T) {
	c := newClient(t)
	initialize(c)
	// The error is on line 3 (index 2), indented by a tab: an operator with
	// nothing to operate on.
	c.open(testURI, "a = 1\nb = 2\n\tc = *\nd = 4\n")
	msgs := c.run()

	n := findNotification(msgs, "textDocument/publishDiagnostics")
	if n == nil {
		t.Fatal("no diagnostics published on didOpen")
	}
	params := n["params"].(map[string]any)
	if params["uri"] != testURI {
		t.Errorf("uri = %v, want %v", params["uri"], testURI)
	}
	diags := params["diagnostics"].([]any)
	if len(diags) == 0 {
		t.Fatal("a file that does not parse produced no diagnostic")
	}

	d := diags[0].(map[string]any)
	rng := d["range"].(map[string]any)
	start := rng["start"].(map[string]any)
	if line := int(start["line"].(float64)); line != 2 {
		t.Errorf("diagnostic on line %d, want 2", line)
	}
	// The indentation must be counted back in, not swallowed: tauerr counts
	// from the first non blank character, the editor counts from the margin.
	// "\tc = *" puts the '*' at character 5.
	if ch := int(start["character"].(float64)); ch != 5 {
		t.Errorf("character = %d, want 5 (the tab must be counted)", ch)
	}
	if d["message"] == "" {
		t.Error("diagnostic without a message")
	}
	if strings.Contains(d["message"].(string), "\n") {
		t.Errorf("message should be one line, got %q", d["message"])
	}
}

func TestCleanFileHasNoDiagnostics(t *testing.T) {
	c := newClient(t)
	initialize(c)
	c.open(testURI, "greet = fn(name) {\n\tprintln(\"hello {name}\")\n}\ngreet(\"world\")\n")
	msgs := c.run()

	n := findNotification(msgs, "textDocument/publishDiagnostics")
	if n == nil {
		t.Fatal("no diagnostics notification")
	}
	diags := n["params"].(map[string]any)["diagnostics"].([]any)
	if len(diags) != 0 {
		t.Fatalf("a file that parses produced %d diagnostics: %v", len(diags), diags)
	}
}

func TestDiagnosticsFollowChanges(t *testing.T) {
	c := newClient(t)
	initialize(c)
	c.open(testURI, "a = 1\n")
	c.notify("textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": testURI, "version": 2},
		"contentChanges": []any{map[string]any{"text": "a = 1\nb = (\n"}},
	})
	msgs := c.run()

	var last []any
	for _, m := range msgs {
		if m["method"] == "textDocument/publishDiagnostics" {
			last = m["params"].(map[string]any)["diagnostics"].([]any)
		}
	}
	if len(last) == 0 {
		t.Fatal("didChange did not republish diagnostics for the broken buffer")
	}
}

func TestCompletionOffersBuiltinsKeywordsAndLocalNames(t *testing.T) {
	c := newClient(t)
	initialize(c)
	c.open(testURI, "answer = 42\ndouble = fn(x) { x * 2 }\n\n")
	id := c.request("textDocument/completion", map[string]any{
		"textDocument": doc(testURI),
		"position":     at(2, 0),
	})
	msgs := c.run()

	res, ok := findResponse(t, msgs, id)["result"].(map[string]any)
	if !ok {
		t.Fatal("completion returned no result")
	}
	labels := map[string]map[string]any{}
	for _, it := range res["items"].([]any) {
		m := it.(map[string]any)
		labels[m["label"].(string)] = m
	}
	for _, want := range []string{"answer", "double", "println", "len", "fn", "if", "import"} {
		if _, ok := labels[want]; !ok {
			t.Errorf("completion is missing %q", want)
		}
	}
	if got := labels["double"]["detail"]; got != "fn(x)" {
		t.Errorf("detail of double = %v, want fn(x)", got)
	}
	if labels["println"]["documentation"] == "" {
		t.Error("builtin completion carries no documentation")
	}
}

// writeModule puts a small module on disk next to a source file, so that the
// import can be resolved the way the runtime resolves it.
func writeModule(t *testing.T) (mainPath, mainURI string) {
	t.Helper()
	dir := t.TempDir()

	mod := "# Greet writes a greeting for name.\nGreet = fn(name) {\n\tprintln(\"hi {name}\")\n}\n\n# Version is the module's version.\nVersion = \"1.0\"\n\nhidden = 7\n"
	if err := os.WriteFile(filepath.Join(dir, "greetings.tau"), []byte(mod), 0644); err != nil {
		t.Fatal(err)
	}
	mainPath = filepath.Join(dir, "main.tau")
	return mainPath, pathToURI(mainPath)
}

func TestCompletionAfterDotListsExportedModuleNames(t *testing.T) {
	_, uri := writeModule(t)

	c := newClient(t)
	initialize(c)
	src := "g = import(\"greetings\")\ng.\n"
	c.open(uri, src)
	id := c.request("textDocument/completion", map[string]any{
		"textDocument": doc(uri),
		"position":     at(1, 2),
	})
	msgs := c.run()

	res := findResponse(t, msgs, id)["result"].(map[string]any)
	labels := map[string]bool{}
	for _, it := range res["items"].([]any) {
		labels[it.(map[string]any)["label"].(string)] = true
	}
	if !labels["Greet"] || !labels["Version"] {
		t.Errorf("module completion = %v, want Greet and Version", labels)
	}
	if labels["hidden"] {
		t.Error("an unexported name leaked into completion")
	}
	if labels["println"] {
		t.Error("builtins should not be offered after a dot")
	}
}

func TestHover(t *testing.T) {
	_, uri := writeModule(t)

	c := newClient(t)
	initialize(c)
	c.open(uri, "g = import(\"greetings\")\ntotal = 3\nprintln(total)\ng.Greet(\"a\")\n")

	builtin := c.request("textDocument/hover", map[string]any{
		"textDocument": doc(uri), "position": at(2, 2),
	})
	local := c.request("textDocument/hover", map[string]any{
		"textDocument": doc(uri), "position": at(2, 10),
	})
	member := c.request("textDocument/hover", map[string]any{
		"textDocument": doc(uri), "position": at(3, 4),
	})
	msgs := c.run()

	text := func(id int) string {
		r, ok := findResponse(t, msgs, id)["result"].(map[string]any)
		if !ok {
			return ""
		}
		return r["contents"].(map[string]any)["value"].(string)
	}

	if got := text(builtin); !strings.Contains(got, "println(...)") {
		t.Errorf("hover on a builtin = %q, want its signature", got)
	}
	if got := text(local); !strings.Contains(got, "line 2") {
		t.Errorf("hover on a local name = %q, want the line it was defined on", got)
	}
	if got := text(member); !strings.Contains(got, "Greet writes a greeting") {
		t.Errorf("hover on a module member = %q, want its doc comment", got)
	}
}

func TestDefinition(t *testing.T) {
	_, uri := writeModule(t)

	c := newClient(t)
	initialize(c)
	c.open(uri, "g = import(\"greetings\")\ncount = 1\nprintln(count)\ng.Greet(\"a\")\n")

	local := c.request("textDocument/definition", map[string]any{
		"textDocument": doc(uri), "position": at(2, 10),
	})
	member := c.request("textDocument/definition", map[string]any{
		"textDocument": doc(uri), "position": at(3, 4),
	})
	module := c.request("textDocument/definition", map[string]any{
		"textDocument": doc(uri), "position": at(0, 0),
	})
	msgs := c.run()

	r, ok := findResponse(t, msgs, local)["result"].(map[string]any)
	if !ok {
		t.Fatal("no definition for a local name")
	}
	if r["uri"] != uri {
		t.Errorf("uri = %v, want %v", r["uri"], uri)
	}
	start := r["range"].(map[string]any)["start"].(map[string]any)
	if int(start["line"].(float64)) != 1 {
		t.Errorf("definition on line %v, want 1", start["line"])
	}

	r, ok = findResponse(t, msgs, member)["result"].(map[string]any)
	if !ok {
		t.Fatal("no definition for a module member")
	}
	if !strings.HasSuffix(r["uri"].(string), "greetings.tau") {
		t.Errorf("member definition points at %v", r["uri"])
	}
	if line := int(r["range"].(map[string]any)["start"].(map[string]any)["line"].(float64)); line != 1 {
		t.Errorf("Greet found on line %d, want 1", line)
	}

	r, ok = findResponse(t, msgs, module)["result"].(map[string]any)
	if !ok {
		t.Fatal("no definition for an import")
	}
	if !strings.HasSuffix(r["uri"].(string), "greetings.tau") {
		t.Errorf("import definition points at %v", r["uri"])
	}
}

func TestDocumentSymbol(t *testing.T) {
	c := newClient(t)
	initialize(c)
	c.open(testURI, "x = 1\nf = fn(a, b) {\n\tinner = a + b\n\tinner\n}\nm = import(\"strings\")\n")
	id := c.request("textDocument/documentSymbol", map[string]any{"textDocument": doc(testURI)})
	msgs := c.run()

	res, ok := findResponse(t, msgs, id)["result"].([]any)
	if !ok {
		t.Fatal("documentSymbol returned no list")
	}
	kinds := map[string]int{}
	for _, it := range res {
		m := it.(map[string]any)
		kinds[m["name"].(string)] = int(m["kind"].(float64))
	}
	if len(kinds) != 3 {
		t.Fatalf("symbols = %v, want exactly the three top level names", kinds)
	}
	if kinds["f"] != int(kindFunction) {
		t.Errorf("f has kind %d, want function", kinds["f"])
	}
	if kinds["m"] != int(kindModule) {
		t.Errorf("m has kind %d, want module", kinds["m"])
	}
	if _, ok := kinds["inner"]; ok {
		t.Error("a name defined inside a function is not a document symbol")
	}
}

func TestFormatting(t *testing.T) {
	c := newClient(t)
	initialize(c)
	c.open(testURI, "x   =   1\nf = fn(a){a+1}\n")
	id := c.request("textDocument/formatting", map[string]any{
		"textDocument": doc(testURI),
		"options":      map[string]any{"tabSize": 4, "insertSpaces": false},
	})
	msgs := c.run()

	res, ok := findResponse(t, msgs, id)["result"].([]any)
	if !ok || len(res) == 0 {
		t.Fatal("formatting returned no edit for a badly spaced file")
	}
	edit := res[0].(map[string]any)
	got := edit["newText"].(string)
	if strings.Contains(got, "x   =") {
		t.Errorf("the file was not formatted: %q", got)
	}
	if !strings.Contains(got, "x = 1") {
		t.Errorf("formatted text = %q", got)
	}
}

func TestFormattingBrokenFileReturnsNull(t *testing.T) {
	c := newClient(t)
	initialize(c)
	c.open(testURI, "x = (\n")
	id := c.request("textDocument/formatting", map[string]any{
		"textDocument": doc(testURI),
	})
	msgs := c.run()

	if r := findResponse(t, msgs, id); r["result"] != nil {
		t.Errorf("formatting a broken file returned %v, want null", r["result"])
	}
}

func TestDidCloseClearsDiagnostics(t *testing.T) {
	c := newClient(t)
	initialize(c)
	c.open(testURI, "x = (\n")
	c.notify("textDocument/didClose", map[string]any{"textDocument": doc(testURI)})
	msgs := c.run()

	var last []any
	for _, m := range msgs {
		if m["method"] == "textDocument/publishDiagnostics" {
			last = m["params"].(map[string]any)["diagnostics"].([]any)
		}
	}
	if len(last) != 0 {
		t.Errorf("diagnostics not cleared on close: %v", last)
	}
}

func TestShutdownThenExit(t *testing.T) {
	c := newClient(t)
	initialize(c)
	id := c.request("shutdown", nil)
	msgs := c.run()

	r := findResponse(t, msgs, id)
	if _, isErr := r["error"]; isErr {
		t.Fatalf("shutdown answered with an error: %v", r)
	}
	if r["result"] != nil {
		t.Errorf("shutdown result = %v, want null", r["result"])
	}
}

// UTF-16 is what the protocol counts in, and tau strings hold whatever the
// author typed, so a position past an emoji has to survive the round trip.
func TestPositionsAreUTF16(t *testing.T) {
	d := newDocument("file:///x.tau", "s = \"🎉\"\nx = 1\n", 1)

	off := strings.Index(d.text, "🎉")
	p := d.position(off)
	if p.Line != 0 || p.Character != 5 {
		t.Errorf("position = %+v, want line 0 character 5", p)
	}
	// The emoji is two code units, so what follows it starts at 7.
	if got := d.position(off + len("🎉")); got.Character != 7 {
		t.Errorf("character after the emoji = %d, want 7", got.Character)
	}
	if got := d.offset(position{Line: 0, Character: 7}); got != off+len("🎉") {
		t.Errorf("offset = %d, want %d", got, off+len("🎉"))
	}
	if got := d.offset(position{Line: 1, Character: 0}); got != strings.Index(d.text, "x = 1") {
		t.Errorf("offset of line 1 = %d", got)
	}
}

func TestOutOfRangePositionsAreClamped(t *testing.T) {
	d := newDocument("file:///x.tau", "a = 1\n", 1)
	for _, p := range []position{
		{Line: -1, Character: -1},
		{Line: 99, Character: 99},
		{Line: 0, Character: 500},
	} {
		if off := d.offset(p); off < 0 || off > len(d.text) {
			t.Errorf("offset(%+v) = %d, outside the buffer", p, off)
		}
	}
}
