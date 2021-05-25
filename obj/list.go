package obj

import (
	"fmt"
	"strconv"
	"strings"
)

type List struct {
	l []Object
	*Env
}

func NewList(elems ...Object) Object {
	return &List{append([]Object{}, elems...), NewEnv()}
}

func (l List) Type() Type {
	return LIST
}

func (l List) Len() int {
	return len(l.l)
}

func (l List) Val(i int64) Object {
	return l.l[i]
}

func (l List) String() string {
	var elements []string

	for _, e := range l.l {
		if e.Type() == STRING {
			elements = append(elements, strconv.Quote(e.String()))
		} else {
			elements = append(elements, e.String())
		}
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}
