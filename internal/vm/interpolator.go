package vm

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/NicoNex/tau/internal/compiler"
	"github.com/NicoNex/tau/internal/parser"
)

const eof = -1

type interpolator struct {
	s       string
	pos     int
	width   int
	nblocks int
	state   *State
	strings.Builder
}

func newInterpolator(s string, state *State) interpolator {
	return interpolator{s: s, state: state}
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
			tree, errs := parser.Parse(s)
			if len(errs) > 0 {
				return "", parserError("interpolation errors:", errs)
			}

			// compile the code
			c := compiler.NewWithState(i.state.Symbols, &i.state.Consts)
			if err := c.Compile(tree); err != nil {
				return "", err
			}

			// run the code
			tvm := NewWithState(c.Bytecode(), i.state)
			if err := tvm.Run(); err != nil {
				return "", err
			}

			// write in the buffer the resulting object
			o := tvm.LastPoppedStackElem()
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
