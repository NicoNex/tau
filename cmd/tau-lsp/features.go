package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// LSP CompletionItemKind values used here.
const (
	itemKindFunction = 3
	itemKindVariable = 6
	itemKindModule   = 9
	itemKindKeyword  = 14
)

type completionItem struct {
	Label         string `json:"label"`
	Kind          int    `json:"kind"`
	Detail        string `json:"detail,omitempty"`
	Documentation string `json:"documentation,omitempty"`
	SortText      string `json:"sortText,omitempty"`
}

// qualifier is the module name written before the dot at the cursor, if the
// cursor sits in a `mod.member` expression, and the part of the member
// already typed.
func qualifier(doc *document, off int) (mod, prefix string) {
	// Walk back over what has been typed of the member.
	i := off
	for i > 0 && isWordByte(doc.text[i-1]) {
		i--
	}
	prefix = doc.text[i:off]

	if i == 0 || doc.text[i-1] != '.' {
		return "", prefix
	}
	j := i - 1
	for j > 0 && isWordByte(doc.text[j-1]) {
		j--
	}
	return doc.text[j : i-1], prefix
}

func (s *server) completion(params json.RawMessage) any {
	var p textDocumentPositionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return emptyCompletion()
	}
	doc := s.document(p.TextDocument.URI)
	if doc == nil {
		return emptyCompletion()
	}

	off := doc.offset(p.Position)
	info := doc.info()
	mod, _ := qualifier(doc, off)

	var items []completionItem

	// After `mod.` only that module's exported names make sense.
	if mod != "" {
		path, ok := info.imports[mod]
		if !ok {
			return emptyCompletion()
		}
		minfo, _ := moduleInfo(path)
		for _, m := range exported(minfo) {
			items = append(items, completionItem{
				Label:         m.name,
				Kind:          kindToItemKind(m.kind),
				Detail:        m.detail,
				Documentation: m.doc,
			})
		}
		sortItems(items)
		return map[string]any{"isIncomplete": false, "items": items}
	}

	// The names of this file first, they are the ones being written about.
	for _, sym := range info.symbols {
		items = append(items, completionItem{
			Label:         sym.name,
			Kind:          kindToItemKind(sym.kind),
			Detail:        sym.detail,
			Documentation: sym.doc,
			SortText:      "0" + sym.name,
		})
	}
	for _, b := range builtins {
		d := builtinDocs[b]
		items = append(items, completionItem{
			Label:         b,
			Kind:          itemKindFunction,
			Detail:        d.signature,
			Documentation: d.summary,
			SortText:      "1" + b,
		})
	}
	for _, k := range keywords {
		items = append(items, completionItem{
			Label:    k,
			Kind:     itemKindKeyword,
			SortText: "2" + k,
		})
	}

	return map[string]any{"isIncomplete": false, "items": items}
}

func emptyCompletion() any {
	return map[string]any{"isIncomplete": false, "items": []completionItem{}}
}

func sortItems(items []completionItem) {
	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
}

func kindToItemKind(k symbolKind) int {
	switch k {
	case kindFunction:
		return itemKindFunction
	case kindModule:
		return itemKindModule
	default:
		return itemKindVariable
	}
}

/* =========================
   Hover
   ========================= */

// wordAt returns the identifier under the offset and its bounds.
func wordAt(doc *document, off int) (string, int, int) {
	if off > len(doc.text) {
		off = len(doc.text)
	}
	start := off
	for start > 0 && isWordByte(doc.text[start-1]) {
		start--
	}
	end := off
	for end < len(doc.text) && isWordByte(doc.text[end]) {
		end++
	}
	return doc.text[start:end], start, end
}

func (s *server) hover(params json.RawMessage) any {
	var p textDocumentPositionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil
	}
	doc := s.document(p.TextDocument.URI)
	if doc == nil {
		return nil
	}

	off := doc.offset(p.Position)
	word, start, end := wordAt(doc, off)
	if word == "" {
		return nil
	}
	mod, _ := qualifier(doc, end)

	value := hoverText(doc, mod, word)
	if value == "" {
		return nil
	}
	return map[string]any{
		"contents": map[string]any{"kind": "markdown", "value": value},
		"range":    doc.rangeOf(start, end),
	}
}

func hoverText(doc *document, mod, word string) string {
	info := doc.info()

	// A member of an imported module: its own doc comment, read where it is
	// written.
	if mod != "" {
		path, ok := info.imports[mod]
		if !ok {
			return ""
		}
		minfo, _ := moduleInfo(path)
		if minfo == nil {
			return ""
		}
		sym, ok := minfo.byName[word]
		if !ok {
			return ""
		}
		return markdown(signature(*sym), sym.doc, fmt.Sprintf("defined in `%s`", path))
	}

	if sym, ok := info.byName[word]; ok {
		where := fmt.Sprintf("defined on line %d", lineOf(doc.text, sym.pos)+1)
		if sym.kind == kindModule {
			if sym.module != "" {
				where = fmt.Sprintf("module `%s` at `%s`", sym.detail, sym.module)
			} else {
				where = fmt.Sprintf("module `%s` (not found)", sym.detail)
			}
		}
		return markdown(signature(*sym), sym.doc, where)
	}

	if d, ok := builtinDocs[word]; ok {
		return markdown(d.signature, d.summary, "builtin")
	}
	if keywordSet[word] {
		return markdown(word, "", "keyword")
	}
	return ""
}

// signature is the one line shown at the top of a hover.
func signature(s symbol) string {
	switch s.kind {
	case kindFunction:
		return s.name + " = " + s.detail
	case kindModule:
		return s.name + " = import(\"" + s.detail + "\")"
	default:
		return s.name
	}
}

func markdown(sig, doc, note string) string {
	var b strings.Builder
	if sig != "" {
		b.WriteString("```tau\n")
		b.WriteString(sig)
		b.WriteString("\n```\n")
	}
	if doc != "" {
		b.WriteString("\n")
		b.WriteString(doc)
		b.WriteString("\n")
	}
	if note != "" {
		b.WriteString("\n_")
		b.WriteString(note)
		b.WriteString("_")
	}
	return b.String()
}

/* =========================
   Definition
   ========================= */

func (s *server) definition(params json.RawMessage) any {
	var p textDocumentPositionParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil
	}
	doc := s.document(p.TextDocument.URI)
	if doc == nil {
		return nil
	}

	off := doc.offset(p.Position)
	word, _, end := wordAt(doc, off)
	if word == "" {
		return nil
	}
	info := doc.info()

	// A member of a module: jump into the module's file.
	if mod, _ := qualifier(doc, end); mod != "" {
		path, ok := info.imports[mod]
		if !ok {
			return nil
		}
		minfo, msrc := moduleInfo(path)
		if minfo == nil {
			return nil
		}
		sym, ok := minfo.byName[word]
		if !ok {
			return nil
		}
		return locationIn(path, msrc, sym.pos, sym.end)
	}

	sym, ok := info.byName[word]
	if !ok {
		return nil
	}
	// The name of an import stands for the file it loads.
	if sym.kind == kindModule && sym.module != "" {
		return map[string]any{
			"uri":   pathToURI(sym.module),
			"range": textRange{},
		}
	}
	return map[string]any{
		"uri":   doc.uri,
		"range": doc.rangeOf(sym.pos, sym.end),
	}
}

// locationIn builds a location inside a file the server has not opened, so
// the positions have to be worked out from its source.
func locationIn(path, src string, start, end int) any {
	d := newDocument(pathToURI(path), src, 0)
	return map[string]any{
		"uri":   d.uri,
		"range": d.rangeOf(start, end),
	}
}

/* =========================
   Document symbols
   ========================= */

func (s *server) documentSymbol(params json.RawMessage) any {
	var p struct {
		TextDocument textDocumentIdentifier `json:"textDocument"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		return []any{}
	}
	doc := s.document(p.TextDocument.URI)
	if doc == nil {
		return []any{}
	}

	info := doc.info()
	out := make([]map[string]any, 0, len(info.symbols))
	for _, sym := range info.symbols {
		r := doc.rangeOf(sym.pos, sym.end)
		out = append(out, map[string]any{
			"name":           sym.name,
			"kind":           int(sym.kind),
			"detail":         sym.detail,
			"range":          r,
			"selectionRange": r,
		})
	}
	return out
}
