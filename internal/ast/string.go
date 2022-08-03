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

type String string

func NewString(s string) (Node, error) {
	str, err := escape(s)
	return String(str), err
}

func (s String) Eval(env *obj.Env) obj.Object {
	return obj.NewString(string(s))
}

func (s String) String() string {
	return string(s)
}

func (s String) Quoted() string {
	return strconv.Quote(string(s))
}

func (s String) Compile(c *compiler.Compiler) (position int, err error) {
	return c.Emit(code.OpConstant, c.AddConstant(obj.NewString(string(s)))), nil
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
