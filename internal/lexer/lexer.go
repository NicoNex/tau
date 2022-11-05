package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/NicoNex/tau/internal/item"
)

type lexer struct {
	items chan item.Item
	input string
	start int
	pos   int
	width int
}

type stateFn func(*lexer) stateFn

const eof = -1

func (l *lexer) next() rune {
	var r rune
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += l.width
	return r
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) curr() rune {
	l.backup()
	return l.next()
}

// Consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// Consumes all the runes if they're in the valid set.
func (l *lexer) acceptRun(valid string) bool {
	for strings.IndexRune(valid, l.next()) >= 0 {

	}
	l.backup()
	return true
}

func (l *lexer) acceptUntil(end rune) {
	for cur := l.next(); cur != end && cur != eof; cur = l.next() {
	}
	l.backup()
}

func (l *lexer) emit(t item.Type) {
	l.items <- item.Item{
		Typ: t,
		Val: l.input[l.start:l.pos],
		Pos: l.start,
	}
	l.start = l.pos
}

func (l *lexer) current() string {
	return l.input[l.start:l.pos]
}

func (l *lexer) ignoreSpaces() {
	l.acceptRun(" \n\t\r")
	l.ignore()
}

func (l *lexer) errorf(format string, args ...any) {
	l.items <- item.Item{
		Typ: item.Error,
		Val: fmt.Sprintf(format, args...),
		Pos: l.start,
	}
	l.start = l.pos
}

func (l *lexer) run() {
	l.ignoreSpaces()
	for state := lexExpression; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func lexIdentifier(l *lexer) stateFn {
	var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	if l.acceptRun(chars) {
		l.emit(item.Lookup(l.current()))
	}
	return lexExpression
}

func lexNumber(l *lexer) stateFn {
	var typ = item.Int
	var digits = "0123456789"

	// Optional leading sign
	l.accept("+-")

	// Is it hex?
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}

	l.acceptRun(digits)
	if l.accept(".") {
		typ = item.Float
		l.acceptRun(digits)
	}

	if l.accept("eE") {
		typ = item.Float
		l.accept("+-")
		l.accept("0123456789")
	}

	l.emit(typ)
	return lexExpression
}

func lexString(l *lexer) stateFn {
Loop:
	for {
		switch l.next() {
		case '\\':
			if r := l.next(); r != eof && r != '\n' {
				break
			}
			fallthrough

		case eof, '\n':
			l.errorf("unterminated quoted string")
			return nil

		case '"':
			l.backup()
			break Loop
		}
	}
	l.emit(item.String)
	l.next()
	l.ignore()
	return lexExpression
}

func lexRawString(l *lexer) stateFn {
	if l.peek() == '`' {
		l.emit(item.RawString)
		l.next()
		l.ignore()
		return lexExpression
	}
	l.next()
	return lexRawString
}

func lexPlus(l *lexer) stateFn {
	switch l.next() {
	case '=':
		l.emit(item.PlusAssign)

	case '+':
		l.emit(item.PlusPlus)

	default:
		l.backup()
		l.emit(item.Plus)
	}
	return lexExpression
}

func lexMinus(l *lexer) stateFn {
	switch l.next() {
	case '=':
		l.emit(item.MinusAssign)

	case '-':
		l.emit(item.MinusMinus)

	default:
		l.backup()
		l.emit(item.Minus)
	}
	return lexExpression
}

func lexTimes(l *lexer) stateFn {
	switch l.next() {
	case '=':
		l.emit(item.AsteriskAssign)

	case '*':
		l.emit(item.Power)

	default:
		l.backup()
		l.emit(item.Asterisk)
	}
	return lexExpression
}

func lexSlash(l *lexer) stateFn {
	switch l.next() {
	case '=':
		l.emit(item.SlashAssign)

	default:
		l.backup()
		l.emit(item.Slash)
	}
	return lexExpression
}

func lexMod(l *lexer) stateFn {
	switch l.next() {
	case '=':
		l.emit(item.ModulusAssign)

	default:
		l.backup()
		l.emit(item.Modulus)
	}
	return lexExpression
}

func lexExpression(l *lexer) stateFn {
	switch r := l.next(); {

	case isSpace(r):
		l.ignore()

	case isLetter(r):
		l.backup()
		return lexIdentifier

	case r == '\n':
		l.emit(item.Semicolon)
		l.ignoreSpaces()

	case r == '"':
		l.ignore()
		return lexString

	case r == '`':
		l.ignore()
		return lexRawString

	case r == ';':
		l.emit(item.Semicolon)
		l.ignoreSpaces()

	case r == ':':
		l.emit(item.Colon)

	case r == '(':
		l.emit(item.LParen)
		l.ignoreSpaces()

	case r == ')':
		l.emit(item.RParen)

	case r == '[':
		l.emit(item.LBracket)
		l.ignoreSpaces()

	case r == ']':
		l.emit(item.RBracket)

	case r == ',':
		l.emit(item.Comma)
		l.ignoreSpaces()

	case r == '.':
		l.emit(item.Dot)
		l.ignoreSpaces()

	case r == '{':
		l.emit(item.LBrace)
		l.ignoreSpaces()

	case r == '}':
		l.emit(item.RBrace)

	case r == '+':
		return lexPlus

	case r == '-':
		return lexMinus

	case r == '*':
		return lexTimes

	case r == '/':
		return lexSlash

	case r == '%':
		return lexMod

	case r == '=':
		if l.next() == '=' {
			l.emit(item.Equals)
		} else {
			l.backup()
			l.emit(item.Assign)
		}

	case r == '!':
		if l.next() == '=' {
			l.emit(item.NotEquals)
		} else {
			l.backup()
			l.emit(item.Bang)
		}

	case r == '~':
		l.emit(item.BwNot)

	case r == '<':
		next := l.next()
		if next == '=' {
			l.emit(item.LTEQ)
		} else if next == '<' {
			if l.next() == '=' {
				l.emit(item.LShiftAssign)
			} else {
				l.backup()
				l.emit(item.LShift)
			}
		} else {
			l.backup()
			l.emit(item.LT)
		}

	case r == '>':
		next := l.next()
		if next == '=' {
			l.emit(item.GTEQ)
		} else if next == '>' {
			if l.next() == '=' {
				l.emit(item.RShiftAssign)
			} else {
				l.backup()
				l.emit(item.RShift)
			}
		} else {
			l.backup()
			l.emit(item.GT)
		}

	case r == '&':
		next := l.next()
		if next == '&' {
			l.emit(item.And)
		} else if next == '=' {
			l.emit(item.BwAndAssign)
		} else {
			l.backup()
			l.emit(item.BwAnd)
		}

	case r == '|':
		next := l.next()
		if next == '|' {
			l.emit(item.Or)
		} else if next == '=' {
			l.emit(item.BwOrAssign)
		} else {
			l.backup()
			l.emit(item.BwOr)
		}

	case r == '^':
		if l.next() == '=' {
			l.emit(item.BwXorAssign)
		} else {
			l.emit(item.BwXor)
		}

	case r == '#':
		l.acceptUntil('\n')
		l.ignoreSpaces()

	case r == eof:
		l.emit(item.EOF)
		return nil

	default:
		if isNumber(r) {
			l.backup()
			return lexNumber
		}
		l.errorf("lexer: invalid item %q", r)
	}
	return lexExpression
}

func isLetter(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\r'
}

func isNumber(r rune) bool {
	return r == '+' || r == '-' || unicode.IsNumber(r)
}

func Lex(in string) chan item.Item {
	l := &lexer{
		input: in,
		items: make(chan item.Item),
	}
	go l.run()
	return l.items
}
