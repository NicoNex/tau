package obj

import "fmt"

type Error string

func NewError(f string, a ...interface{}) Object {
	var ret = Error(fmt.Sprintf(f, a...))
	return &ret
}

func (e Error) Type() Type {
	return ERROR
}

func (e Error) String() string {
	return fmt.Sprintf("error: %s", string(e))
}
