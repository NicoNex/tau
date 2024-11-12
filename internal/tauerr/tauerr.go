package tauerr

import "C"

import (
	"fmt"
	"strings"
)

type TauErr struct {
	Line    int
	Column  int
	File    string
	Message string
}

func (e TauErr) Error() string {
	return e.Message
}

func New(file, input string, pos int, s string, a ...any) error {
	if file == "" {
		file = "<stdin>"
	}

	line, lineno, rel := line(input, pos)
	return TauErr{
		File:   file,
		Line:   lineno,
		Column: rel,
		Message: fmt.Sprintf(
			"error in file %s at line %d:\n    %s\n    %s\n%s",
			file,
			lineno,
			line,
			arrow(rel),
			fmt.Sprintf(s, a...),
		),
	}
}

func NewFromBookmark(file string, b Bookmark, s string, a ...any) error {
	if b == (Bookmark{}) {
		return fmt.Errorf(s, a...)
	}

	return TauErr{
		Line:   int(b.lineno),
		Column: int(b.pos),
		File:   file,
		Message: fmt.Sprintf(
			"error in file %s at line %d:\n    %s\n    %s\n%s",
			file,
			int(b.lineno),
			C.GoString(b.line),
			arrow(int(b.pos)),
			fmt.Sprintf(s, a...),
		),
	}
}

func line(input string, pos int) (line string, lineno, relative int) {
	s, e := start(input, pos), end(input, int(pos))
	l := input[s:e]
	line = strings.TrimLeft(l, " \t")
	return line, lineNo(input, pos), len(line) - (e - pos)
}

func start(s string, pos int) int {
	for i := pos - 1; i >= 0; i-- {
		if s[i] == '\n' {
			return i + 1
		}
	}
	return 0
}

func end(s string, pos int) int {
	for i := pos; i < len(s); i++ {
		if s[i] == '\n' {
			return i
		}
	}
	return len(s)
}

func lineNo(s string, pos int) int {
	var cnt = 1

	for _, b := range s[:pos] {
		if b == '\n' {
			cnt++
		}
	}

	return cnt
}

func arrow(pos int) string {
	var s = make([]byte, pos+1)

	for i := range s {
		if i == pos {
			s[i] = '^'
		} else {
			s[i] = ' '
		}
	}
	return string(s)
}
