package compiler

import "fmt"

type CompilerError struct {
	pos int
	msg string
}

func NewError(pos int, s string, a ...any) *CompilerError {
	return &CompilerError{
		pos: pos,
		msg: fmt.Sprintf(s, a...),
	}
}

func (c *CompilerError) Error() string {
	return c.msg
}

func (c *CompilerError) Pos() int {
	return c.pos
}
