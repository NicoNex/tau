package obj

import (
	"fmt"
	"strings"
)

type Class struct {
	*Env
}

func NewClass() Object {
	return Class{NewEnv()}
}

func (c Class) Type() Type {
	return ObjectType
}

func (c Class) String() string {
	var buf strings.Builder
	buf.WriteString("{")

	i := 0
	for k, v := range c.Env.Store {
		if i < len(c.Env.Store)-1 {
			buf.WriteString(fmt.Sprintf("%s: %s, ", k, v))
		} else {
			buf.WriteString(fmt.Sprintf("%s: %s", k, v))
		}
		i++
	}
	buf.WriteString("}")
	return buf.String()
}
