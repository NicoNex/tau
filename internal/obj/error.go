package obj

import "fmt"

type Error string

func NewError(f string, a ...any) Object {
	return Error(fmt.Sprintf(f, a...))
}

func (e Error) Type() Type {
	return ErrorType
}

func (e Error) String() string {
	return fmt.Sprintf("error: %s", string(e))
}

func (e Error) Val() string {
	return string(e)
}
