package obj

import (
	"fmt"
	"strings"
)

type TauObject struct {
	Store
}

func NewTauObject() Object {
	return TauObject{NewStore()}
}

func (c TauObject) Type() Type {
	return ObjectType
}

func (c TauObject) String() string {
	var buf strings.Builder
	buf.WriteString("{")

	i := 0
	for k, v := range c.Store {
		if i < len(c.Store)-1 {
			buf.WriteString(fmt.Sprintf("%s: %s, ", k, v))
		} else {
			buf.WriteString(fmt.Sprintf("%s: %s", k, v))
		}
		i++
	}
	buf.WriteString("}")
	return buf.String()
}
