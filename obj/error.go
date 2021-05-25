package obj

import "fmt"

type Error struct {
	e string
	*Env
}

func NewError(f string, a ...interface{}) Object {
	return &Error{
		e: fmt.Sprintf(f, a...),
		Env: NewEnv(),
	}
}

func (e Error) Type() Type {
	return ERROR
}

func (e Error) String() string {
	return fmt.Sprintf("error: %s", e.e)
}
