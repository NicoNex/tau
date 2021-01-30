package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/NicoNex/tau/item"
)

type lexer struct {
	input string
	start int
	pos   int
	width int
	items chan item.Item
}

type stateFn func(*lexer) stateFn

func (l *lexer) next() rune {
	var r rune
	if l.pos >= len(l.input) {
		l.width = 0
		return 0
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
	for l.next() != end {

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

func (l *lexer) errorf(format string, args ...interface{}) {
	l.items <- item.Item{
		Typ: item.ERROR,
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
	var typ = item.INT
	var digits = "0123456789"

	// Optional leading sign
	l.accept("+-")

	// Is it hex?
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}

	l.acceptRun(digits)
	if l.accept(".") {
		typ = item.FLOAT
		l.acceptRun(digits)
	}

	if l.accept("eE") {
		typ = item.FLOAT
		l.accept("+-")
		l.accept("0123456789")
	}

	l.emit(typ)
	return lexExpression
}

func lexString(l *lexer) stateFn {
	if l.peek() == '"' {
		l.emit(item.STRING)
		l.next()
		l.ignore()
		return lexExpression
	}
	l.next()
	return lexString
}

func lexPlus(l *lexer) stateFn {
	switch l.next() {
	case '=':
		l.emit(item.PLUS_ASSIGN)

	case '+':
		l.emit(item.PLUSPLUS)

	default:
		l.backup()
		l.emit(item.PLUS)
	}
	return lexExpression
}

func lexMinus(l *lexer) stateFn {
	switch l.next() {
	case '=':
		l.emit(item.MINUS_ASSIGN)

	case '-':
		l.emit(item.MINUSMINUS)

	default:
		l.backup()
		l.emit(item.MINUS)
	}
	return lexExpression
}

func lexTimes(l *lexer) stateFn {
	switch l.next() {
	case '=':
		l.emit(item.ASTERISK_ASSIGN)

	case '*':
		l.emit(item.POWER)

	default:
		l.backup()
		l.emit(item.ASTERISK)
	}
	return lexExpression
}

func lexSlash(l *lexer) stateFn {
	switch l.next() {
	case '=':
		l.emit(item.SLASH_ASSIGN)

	default:
		l.backup()
		l.emit(item.SLASH)
	}
	return lexExpression
}

func lexMod(l *lexer) stateFn {
	switch l.next() {
	case '=':
		l.emit(item.MODULUS_ASSIGN)

	default:
		l.backup()
		l.emit(item.MODULUS)
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
		l.emit(item.SEMICOLON)
		l.ignoreSpaces()

	case r == '"':
		l.ignore()
		return lexString

	case r == ';':
		l.emit(item.SEMICOLON)
		l.ignoreSpaces()

	case r == '(':
		l.emit(item.LPAREN)

	case r == ')':
		l.emit(item.RPAREN)

	case r == '[':
		l.emit(item.LBRACKET)

	case r == ']':
		l.emit(item.RBRACKET)

	case r == ',':
		l.emit(item.COMMA)
		l.ignoreSpaces()

	case r == '{':
		l.emit(item.LBRACE)
		l.ignoreSpaces()

	case r == '}':
		l.emit(item.RBRACE)

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
			l.emit(item.EQ)
		} else {
			l.backup()
			l.emit(item.ASSIGN)
		}

	case r == '!':
		if l.next() == '=' {
			l.emit(item.NOT_EQ)
		} else {
			l.backup()
			l.emit(item.BANG)
		}

	case r == '<':
		next := l.next()
		if next == '=' {
			l.emit(item.LT_EQ)
		} else if next == '<' {
			if l.next() == '=' {
				l.emit(item.LSHIFT_ASSIGN)
			} else {
				l.backup()
				l.emit(item.LSHIFT)
			}
		} else {
			l.backup()
			l.emit(item.LT)
		}

	case r == '>':
		next := l.next()
		if next == '=' {
			l.emit(item.GT_EQ)
		} else if next == '>' {
			if l.next() == '=' {
				l.emit(item.RSHIFT_ASSIGN)
			} else {
				l.backup()
				l.emit(item.RSHIFT)
			}
		} else {
			l.backup()
			l.emit(item.GT)
		}

	case r == '&':
		next := l.next()
		if next == '&' {
			l.emit(item.AND)
		} else if next == '=' {
			l.emit(item.BWAND_ASSIGN)
		} else {
			l.backup()
			l.emit(item.BWAND)
		}

	case r == '|':
		next := l.next()
		if next == '|' {
			l.emit(item.OR)
		} else if next == '=' {
			l.emit(item.BWOR_ASSIGN)
		} else {
			l.backup()
			l.emit(item.BWOR)
		}

	case r == '^':
		if l.next() == '=' {
			l.emit(item.BWXOR_ASSIGN)
		} else {
			l.emit(item.BWXOR)
		}

	case r == '#':
		l.acceptUntil('\n')
		l.next()
		l.ignore()

	case r == 0:
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
