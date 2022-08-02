package obj

import (
	"fmt"
	"strings"
)

type List []Object

func NewList(elems ...Object) Object {
	return append(List{}, elems...)
}

func (l List) Type() Type {
	return ListType
}

func (l List) String() string {
	var elements []string

	for _, e := range l {
		if s, ok := e.(*String); ok {
			elements = append(elements, s.Quoted())
		} else {
			elements = append(elements, e.String())
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (l List) Val() []Object {
	return []Object(l)
}
