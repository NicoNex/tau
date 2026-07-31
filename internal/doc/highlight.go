package doc

import (
	"html"
	"html/template"
	"strings"

	"github.com/NicoNex/tau/internal/item"
	"github.com/NicoNex/tau/internal/lexer"
	"github.com/NicoNex/tau/internal/obj"
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

	// Where each token starts, worked out before anything is written: a
	// string begins at what is between the quotes rather than at the quote,
	// and the one in front of it has to stop short of the quote it does not
	// own.
	starts := make([]int, len(toks))
	for i, t := range toks {
		p := t.Pos
		if t.Is(item.String) || t.Is(item.RawString) {
			if p > 0 && (src[p-1] == '"' || src[p-1] == '`') {
				p--
			}
		}
		if p < 0 || p > len(src) || (i > 0 && p < starts[i-1]) {
			return plain(src)
		}
		starts[i] = p
	}

	var (
		out strings.Builder
		pos int
	)

	for i := range toks {
		if starts[i] < pos {
			return plain(src)
		}

		// Whatever stood between the last token and this one is spacing, and
		// belongs to nobody.
		out.WriteString(html.EscapeString(src[pos:starts[i]]))

		end := len(src)
		if i+1 < len(toks) {
			end = starts[i+1]
		}
		// The token is what is left once the spacing before the next one is
		// taken off the end.
		text := strings.TrimRight(src[starts[i]:end], " \t\n")

		if class := class(toks, i); class != "" {
			open := `<span class="` + class + `">`
			out.WriteString(open)
			// A newline inside the token ends the span and starts another,
			// so that no line is left holding half of one.
			out.WriteString(strings.ReplaceAll(html.EscapeString(text), "\n", `</span>`+"\n"+open))
			out.WriteString(`</span>`)
		} else {
			out.WriteString(html.EscapeString(text))
		}
		pos = starts[i] + len(text)
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

// builtins are the names that are always in scope, taken from the interpreter
// rather than written out again here: one added there is one painted here,
// with nothing to keep in step by hand.
var builtins = func() map[string]bool {
	m := make(map[string]bool, len(obj.Builtins))
	for _, name := range obj.Builtins {
		m[name] = true
	}
	return m
}()

// class is what a token is painted as.
//
// A name is read in the company it keeps: one with a bracket after it is
// something being called, which is what a reader looks for first when
// following what a piece of code does. The rest of the names, and all the
// punctuation between them, get nothing - colouring everything is colouring
// nothing.
func class(toks []item.Item, at int) string {
	i := toks[at]

	if i.Is(item.Ident) {
		switch {
		case builtins[i.Val]:
			return "c-bi"
		case at+1 < len(toks) && toks[at+1].Is(item.LParen):
			return "c-fn"
		default:
			return ""
		}
	}

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
