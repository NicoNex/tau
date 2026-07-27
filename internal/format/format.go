// Package format writes tau source in the canonical style: tabs for
// indentation, one space around binary operators, none where nothing is
// separated.
//
// The source is formatted from the token stream rather than from the AST,
// for one reason: the AST has no comments, and a formatter that eats the
// comments is not a formatter. The lines the author wrote are kept as they
// are; what changes is the indentation and the spacing inside them.
package format

import (
	"errors"
	"fmt"
	"strings"

	"github.com/NicoNex/tau/internal/item"
	"github.com/NicoNex/tau/internal/lexer"
	"github.com/NicoNex/tau/internal/parser"
)

// Source returns src formatted. A file that doesn't parse comes back
// untouched: there is nothing sensible to do with source nobody can read.
func Source(file, src string) (string, error) {
	if _, errs := parser.Parse(file, src); len(errs) > 0 {
		return "", errs[0]
	}

	out := print(tokens(src))

	// The formatter must not change the program. Lexing the result and
	// comparing it with what went in costs one more pass and turns any
	// mistake into a refusal instead of a broken file.
	if err := sameCode(src, out); err != nil {
		return out, err
	}
	return out, nil
}

// tok is an item together with the line it was written on, which is what
// tells an empty line from a wrapped one.
type tok struct {
	item.Item
	line int
}

// tokens reads the whole stream, dropping the end of file marker.
func tokens(src string) []tok {
	var (
		out  []tok
		line = 1
		pos  int
	)

	for i := range lexer.Lex(src) {
		if i.Is(item.EOF) || i.Is(item.Error) {
			break
		}

		// The lines between the last token and this one, counted on the
		// source itself: the stream carries no whitespace.
		if i.Pos > pos && i.Pos <= len(src) {
			line += strings.Count(src[pos:i.Pos], "\n")
			pos = i.Pos
		}
		out = append(out, tok{Item: i, line: line})
	}

	return out
}

// text is the source of a token, quotes included: the lexer hands over what
// is between the quotes, not the quotes themselves.
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

// isBreak reports whether the token is the newline the lexer turns into a
// semicolon. A semicolon the author wrote is text, a newline is a line break
// and comes back as one.
func (t tok) isBreak() bool {
	return t.Is(item.Semicolon) && t.Val != ";"
}

// opens and closes are the tokens that change the indentation.
func (t tok) opens() bool {
	return t.Is(item.LBrace) || t.Is(item.LBracket) || t.Is(item.LParen)
}

func (t tok) closes() bool {
	return t.Is(item.RBrace) || t.Is(item.RBracket) || t.Is(item.RParen)
}

// value reports whether something ends with this token: a name, a literal, a
// closing bracket. It is what tells a minus that subtracts from one that
// negates, an index from a list, and a block from a map.
func value(t tok) bool {
	switch t.Typ {
	case item.RParen, item.RBracket, item.RBrace, item.Ident, item.Int,
		item.Float, item.String, item.RawString, item.True, item.False,
		item.Null, item.PlusPlus, item.MinusMinus, item.Function, item.Import:
		return true
	default:
		return false
	}
}

// print walks the tokens and writes them back, one source line at a time:
// where the author broke a line, so does the output.
func print(toks []tok) string {
	var (
		out   strings.Builder
		line  []tok
		kinds []bool // for every open brace, whether it holds a block
		depth int
		prev  int
	)

	flush := func() {
		if len(line) == 0 {
			return
		}

		// A line starting with a closing bracket belongs to the level its
		// opening one was on.
		indent := depth
		if line[0].closes() && indent > 0 {
			indent--
		}

		out.WriteString(strings.Repeat("\t", indent))
		out.WriteString(join(line, &kinds))
		out.WriteByte('\n')

		for _, t := range line {
			if t.opens() {
				depth++
			} else if t.closes() && depth > 0 {
				depth--
			}
		}
		line = line[:0]
	}

	for i, t := range toks {
		if i > 0 && t.line > prev {
			flush()

			// One empty line the author left is kept, more than one is not.
			if t.line > prev+1 {
				out.WriteByte('\n')
			}
		}
		prev = t.line

		if !t.isBreak() {
			line = append(line, t)
		}
	}
	flush()

	// The file ends with exactly one newline.
	return strings.TrimRight(out.String(), "\n") + "\n"
}

// join writes the tokens of one line with the spacing each of them wants.
// kinds is the stack of open braces, telling a block from a map literal: it
// outlives the line, since a block rarely fits in one.
func join(line []tok, kinds *[]bool) string {
	var out strings.Builder

	for i, t := range line {
		var prev tok
		if i > 0 {
			prev = line[i-1]
		}

		// A brace is a block when something it could belong to comes before
		// it, and a map literal otherwise: "if x {" and "else {" against
		// "m = {".
		if t.Is(item.LBrace) {
			*kinds = append(*kinds, i > 0 && (value(prev) || prev.Is(item.Else)))
		}

		block := true
		if t.Is(item.LBrace) {
			block = (*kinds)[len(*kinds)-1]
		} else if t.Is(item.RBrace) {
			if n := len(*kinds); n > 0 {
				block = (*kinds)[n-1]
				*kinds = (*kinds)[:n-1]
			}
		} else if n := len(*kinds); n > 0 {
			block = (*kinds)[n-1]
		}

		if i > 0 && space(prev, t, line[:i], block) {
			out.WriteByte(' ')
		}
		out.WriteString(t.text())
	}

	return out.String()
}

// space reports whether the two tokens want a space between them. before is
// what the line holds so far, needed to tell a sign from an operator, and
// block says whether the innermost open brace is a block or a map literal.
func space(prev, cur tok, before []tok, block bool) bool {
	switch {
	// Nothing follows these: an open bracket, a dot, a negation.
	case prev.Is(item.LParen) || prev.Is(item.LBracket) || prev.Is(item.Dot) ||
		prev.Is(item.Bang) || prev.Is(item.BwNot):
		return false

	// Brackets stick to the name in front of them, which makes a call or an
	// index, and stand apart from anything else, which makes a literal.
	case cur.Is(item.LParen) || cur.Is(item.LBracket):
		return !value(prev)

	case cur.Is(item.RParen) || cur.Is(item.RBracket) || cur.Is(item.Comma) ||
		cur.Is(item.Colon) || cur.Is(item.Semicolon):
		return false

	case cur.Is(item.Dot):
		return false

	// A block breathes, a map literal doesn't: "if x { y }" and {"a": 1}.
	case prev.Is(item.LBrace):
		return block && !cur.Is(item.RBrace)

	case cur.Is(item.RBrace):
		return block && !prev.Is(item.LBrace)

	// A semicolon the author wrote separates: "for i = 0; i < n; ++i".
	case prev.Is(item.Semicolon):
		return true

	// ++ and -- stick to what they count.
	case cur.Is(item.PlusPlus) || cur.Is(item.MinusMinus),
		prev.Is(item.PlusPlus) || prev.Is(item.MinusMinus):
		return false

	// A sign is part of the number it is in front of: nothing between them,
	// while the space before it is whatever the token before wants.
	case (prev.Is(item.Minus) || prev.Is(item.Plus)) && len(before) > 1 &&
		!value(before[len(before)-2]):
		return false

	default:
		return true
	}
}

// sameCode reports whether two sources hold the same program, comments
// included: same tokens, same order, same values.
func sameCode(a, b string) error {
	ta, tb := tokens(a), tokens(b)

	if len(ta) != len(tb) {
		return fmt.Errorf("format: %d tokens became %d, refusing to write", len(ta), len(tb))
	}

	for i := range ta {
		if ta[i].Typ != tb[i].Typ || ta[i].Val != tb[i].Val {
			return fmt.Errorf(
				"format: %q became %q at token %d, refusing to write",
				ta[i].Val, tb[i].Val, i,
			)
		}
	}

	if len(ta) == 0 && strings.TrimSpace(b) != "" {
		return errors.New("format: empty source became something, refusing to write")
	}
	return nil
}
