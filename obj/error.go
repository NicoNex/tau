package obj

import "fmt"

type Error string

func NewError(f string, a ...interface{}) Object {
	return &Error(fmt.Sprintf(f, a...))
}

func (e *Error) Type() Type {
	return ERROR
}

func (e *Error) String() string {
	return fmt.Sprintf("error: %s", string(e))
}
