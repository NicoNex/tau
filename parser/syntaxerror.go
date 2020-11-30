package parser

import (
	"fmt"
	"strings"
)

type SyntaxError struct {
	e string
}

func point(p, l int) string {
	var buf strings.Builder

	for i := 0; i < l; i++ {
		if i == p {
			buf.WriteRune('^')
			continue
		}
		buf.WriteRune(' ')
	}
	return buf.String()
}

func NewSyntaxError(reason, input string, pos int) error {
	return SyntaxError{
		fmt.Sprintf("syntax error: %s\n\t%s\n\t%s",
			reason,
			input,
			point(pos, len(input)),
		),
	}
}

func (s SyntaxError) Error() string {
	return s.e
}
