package obj

import (
	"fmt"
	"strings"
)

type List []Object

func NewList(elems ...Object) Object {
	var ret List
	return append(ret, elems...)
}

func (l List) Type() Type {
	return LIST
}

func (l List) String() string {
	var elements []string

	for _, e := range l {
		elements = append(elements, e.String())
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}
