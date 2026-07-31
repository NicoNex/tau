package doc

import (
	"html"
	"html/template"
	"strings"

	"github.com/NicoNex/tau/internal/item"
	"github.com/NicoNex/tau/internal/lexer"
)

// highlight colours a block of tau written inside a comment.
//
// The lexer of the language does the reading, so what the page paints is what
// the interpreter sees and not a second grammar written in a stylesheet that
// would drift from the first one. It happens here rather than in the browser
// for the same reason the page carries its own style: the file is opened from
// disk, and a highlighter fetched from somewhere else never arrives.
//
// An example in a comment is not always a program - half of them trail off
// into "..." - so anything the lexer cannot read comes out as plain text
// rather than as a guess.
func highlight(src string) template.HTML {
	lines := highlightLines(src)

	parts := make([]string, len(lines))
	for i, l := range lines {
		parts[i] = string(l)
	}
	return template.HTML(strings.Join(parts, "\n"))
}

// highlightLines is the same, one entry per line of the source.
//
// A token may hold newlines - a raw string does - so a span is closed at the
// end of every line it crosses and opened again on the next. Each line then
// stands on its own, which is what putting a number beside one needs.
func highlightLines(src string) []template.HTML {
	toks, ok := lex(src)
	if !ok {
		return plain(src)
	}

	var (
		out strings.Builder
		pos int
	)

	for i, t := range toks {
		if t.Pos < pos || t.Pos > len(src) {
			return plain(src)
		}

		// Whatever stood between the last token and this one is spacing, and
		// belongs to nobody.
		out.WriteString(html.EscapeString(src[pos:t.Pos]))

		end := len(src)
		if i+1 < len(toks) {
			end = toks[i+1].Pos
		}
		// The token is what is left once the spacing before the next one is
		// taken off the end.
		text := strings.TrimRight(src[t.Pos:end], " \t\n")

		if class := class(t); class != "" {
			open := `<span class="` + class + `">`
			out.WriteString(open)
			// A newline inside the token ends the span and starts another,
			// so that no line is left holding half of one.
			out.WriteString(strings.ReplaceAll(html.EscapeString(text), "\n", `</span>`+"\n"+open))
			out.WriteString(`</span>`)
		} else {
			out.WriteString(html.EscapeString(text))
		}
		pos = t.Pos + len(text)
	}
	out.WriteString(html.EscapeString(src[pos:]))

	return split(out.String())
}

// plain is the source with nothing painted on it, which is what anything the
// lexer cannot read comes back as.
func plain(src string) []template.HTML {
	return split(html.EscapeString(src))
}

func split(s string) []template.HTML {
	lines := strings.Split(s, "\n")

	out := make([]template.HTML, len(lines))
	for i, l := range lines {
		out[i] = template.HTML(l)
	}
	return out
}

// lex reads the whole stream, and says whether it is tau at all. The channel
// is drained either way: the lexer runs in a routine of its own and a reader
// that walks off leaves it stuck on a send.
func lex(src string) ([]item.Item, bool) {
	var (
		out []item.Item
		ok  = true
	)

	for i := range lexer.Lex(src) {
		if i.Is(item.EOF) {
			break
		}
		if i.Is(item.Error) {
			ok = false
			break
		}
		out = append(out, i)
	}
	return out, ok
}

// class is what a token is painted as. Names and the punctuation between them
// get none: colouring everything is colouring nothing.
func class(i item.Item) string {
	switch i.Typ {
	case item.Comment:
		return "c-com"

	case item.String, item.RawString:
		return "c-str"

	case item.Int, item.Float:
		return "c-num"

	case item.Function, item.For, item.If, item.Else, item.Return, item.Import,
		item.Tau, item.Break, item.Continue:
		return "c-kw"

	case item.True, item.False, item.Null:
		return "c-lit"

	default:
		return ""
	}
}
