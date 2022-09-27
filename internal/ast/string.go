package ast

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/NicoNex/tau/internal/code"
	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/obj"
)

type parseFn func(string) (Node, []string)

type String struct {
	s      string
	parse  parseFn
	substr []Node
}

func NewString(s string, parse parseFn) (Node, error) {
	str, err := escape(s)
	if err != nil {
		return nil, err
	}

	i := newInterpolator(str, parse)
	nodes, str, err := i.nodes()
	return String{s: str, parse: parse, substr: nodes}, err
}

func (s String) Eval(env *obj.Env) obj.Object {
	if len(s.substr) == 0 {
		return obj.NewString(s.s)
	}

	var subs = make([]any, len(s.substr))
	for i, sub := range s.substr {
		subs[i] = sub.Eval(env)
	}

	return obj.NewString(fmt.Sprintf(s.s, subs...))
}

func (s String) String() string {
	return s.s
}

func (s String) Quoted() string {
	return strconv.Quote(s.s)
}

func (s String) Compile(c *compiler.Compiler) (position int, err error) {
	if len(s.substr) == 0 {
		return c.Emit(code.OpConstant, c.AddConstant(obj.NewString(s.s))), nil
	}

	for _, sub := range s.substr {
		if position, err = sub.Compile(c); err != nil {
			return
		}
		c.RemoveLast()
	}

	return c.Emit(code.OpInterpolate, c.AddConstant(obj.NewString(s.s)), len(s.substr)), nil
}

func escape(s string) (string, error) {
	var buf strings.Builder

	for i := 0; i < len(s); {
		r, width := utf8.DecodeRuneInString(s[i:])

		if r == '\\' {
			i += width
			if i < len(s) {
				r, width := utf8.DecodeRuneInString(s[i:])
				esc, err := escapeRune(r)
				if err != nil {
					return "", err
				}
				buf.WriteRune(esc)
				i += width
				continue
			} else {
				return "", errors.New("newline in string")
			}
		} else {
			buf.WriteRune(r)
		}

		i += width
	}

	return buf.String(), nil
}

func escapeRune(r rune) (rune, error) {
	switch r {
	case 'a':
		return '\a', nil
	case 'b':
		return '\b', nil
	case 'f':
		return '\f', nil
	case 'n':
		return '\n', nil
	case 'r':
		return '\r', nil
	case 't':
		return '\t', nil
	case 'v':
		return '\v', nil
	case '\\':
		return '\\', nil
	case '\'':
		return '\'', nil
	case '"':
		return '"', nil
	default:
		return r, fmt.Errorf(`unknown escape "\%c"`, r)
	}
}

func toAnySlice(args []obj.Object) []any {
	var ret = make([]any, len(args))
	for i, a := range args {
		ret[i] = a
	}
	return ret
}

const eof = -1

var errBadInterpolationSyntax = errors.New("bad interpolation syntax")

type interpolator struct {
	s          string
	pos        int
	width      int
	nblocks    int
	inQuotes   bool
	inBacktick bool
	parse      parseFn
	strings.Builder
}

func newInterpolator(s string, parse parseFn) interpolator {
	return interpolator{s: s, parse: parse}
}

func (i *interpolator) next() (r rune) {
	if i.pos >= len(i.s) {
		i.width = 0
		return eof
	}

	r, i.width = utf8.DecodeRuneInString(i.s[i.pos:])
	i.pos += i.width
	return
}

func (i *interpolator) backup() {
	i.pos -= i.width
}

func (i *interpolator) peek() rune {
	r := i.next()
	i.backup()
	return r
}

func (i *interpolator) enterBlock() {
	i.nblocks++
}

func (i *interpolator) exitBlock() {
	i.nblocks--
}

func (i *interpolator) insideBlock() bool {
	return i.nblocks > 0
}

func (i *interpolator) quotes() {
	i.inQuotes = !i.inQuotes
}

func (i *interpolator) backtick() {
	i.inBacktick = !i.inBacktick
}

func (i *interpolator) insideString() bool {
	return i.inQuotes || i.inBacktick
}

func (i *interpolator) acceptUntil(start, end rune) (string, error) {
	var buf strings.Builder

loop:
	for r := i.next(); ; r = i.next() {
		switch r {
		case eof:
			return "", errBadInterpolationSyntax

		case '"':
			i.quotes()

		case '`':
			i.backtick()

		case start:
			if !i.insideString() {
				i.enterBlock()
			}

		case end:
			if !i.insideString() {
				if !i.insideBlock() {
					break loop
				}
				i.exitBlock()
			}
		}

		buf.WriteRune(r)
	}

	return buf.String(), nil
}

func (i *interpolator) nodes() ([]Node, string, error) {
	var nodes []Node

	for r := i.next(); r != eof; r = i.next() {
		if r == '{' {
			if i.peek() == '{' {
				i.next()
				goto tail
			}

			// Get the code between braces
			s, err := i.acceptUntil('{', '}')
			if err != nil {
				return []Node{}, "", err
			} else if s == "" {
				continue
			}

			// Parse the code
			tree, errs := i.parse(s)
			if len(errs) > 0 {
				return []Node{}, "", i.parserError(errs)
			}

			nodes = append(nodes, tree)
			i.WriteString("%v")
			continue
		} else if r == '}' {
			if i.peek() != '}' {
				return []Node{}, "", errBadInterpolationSyntax
			}
			i.next()
		}

	tail:
		i.WriteRune(r)
	}

	return nodes, i.String(), nil
}

func (i *interpolator) parserError(errs []string) error {
	var buf strings.Builder

	buf.WriteString("interpolation errors:\n")
	for _, e := range errs {
		buf.WriteRune('\t')
		buf.WriteString(e)
		buf.WriteRune('\n')
	}

	return errors.New(buf.String())
}
