package obj

import (
	"fmt"
	"strconv"
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
		if e.Type() == STRING {
			elements = append(elements, strconv.Quote(e.String()))
		} else {
			elements = append(elements, e.String())
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}

func (l List) Val() []Object {
	return []Object(l)
}
