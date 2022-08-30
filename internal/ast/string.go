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
	s     string
	parse parseFn
}

func NewString(s string, parse parseFn) (Node, error) {
	str, err := escape(s)
	return String{s: str, parse: parse}, err
}

func (s String) Eval(env *obj.Env) obj.Object {
	i := newInterpolator(s.s, env, s.parse)
	str, err := i.run()
	if err != nil {
		return obj.NewError(err.Error())
	}
	return obj.NewString(str)
}

func (s String) String() string {
	return s.s
}

func (s String) Quoted() string {
	return strconv.Quote(s.s)
}

func (s String) Compile(c *compiler.Compiler) (position int, err error) {
	return c.Emit(code.OpConstant, c.AddConstant(obj.NewString(s.s))), nil
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

const eof = -1

type interpolator struct {
	s       string
	pos     int
	width   int
	nblocks int
	env     *obj.Env
	parse   parseFn
	strings.Builder
}

func newInterpolator(s string, env *obj.Env, parse parseFn) interpolator {
	return interpolator{s: s, env: env, parse: parse}
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

func (i *interpolator) acceptUntil(start, end rune) (string, error) {
	var buf strings.Builder

loop:
	for r := i.next(); ; r = i.next() {
		switch r {
		case eof:
			return "", errors.New("bad interpolation syntax")

		case start:
			i.enterBlock()
			buf.WriteRune(r)

		case end:
			if !i.insideBlock() {
				if p := i.peek(); p == end {
					buf.WriteRune(end)
					i.next()
					continue loop
				}
				break loop
			}
			i.exitBlock()
			fallthrough

		default:
			buf.WriteRune(r)
		}
	}

	return buf.String(), nil
}

func (i *interpolator) run() (string, error) {
	for r := i.next(); r != eof; r = i.next() {
		if r == '{' {
			if r := i.peek(); r == '{' {
				i.next()
				goto tail
			}

			// get the code between braces
			s, err := i.acceptUntil('{', '}')
			if err != nil {
				return "", err
			} else if s == "" {
				continue
			}

			// parse the code
			tree, errs := i.parse(s)
			if len(errs) > 0 {
				return "", i.parserError(errs)
			}

			// execute the code and write the resulting object in the buffer
			o := tree.Eval(i.env)
			i.WriteString(o.String())
			continue
		} else if p := i.peek(); r == '}' && p == '}' {
			i.next()
		}

	tail:
		i.WriteRune(r)
	}

	return i.String(), nil
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
