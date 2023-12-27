package obj

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type TauObject map[string]Object

func NewTauObject() Object {
	return TauObject(make(map[string]Object))
}

func (o TauObject) Type() Type {
	return ObjectType
}

func (o TauObject) String() string {
	var buf strings.Builder
	buf.WriteString("{")

	i := 0
	for k, v := range o {
		if i < len(o)-1 {
			buf.WriteString(fmt.Sprintf("%s: %s, ", k, v))
		} else {
			buf.WriteString(fmt.Sprintf("%s: %s", k, v))
		}
		i++
	}
	buf.WriteString("}")
	return buf.String()
}

func (to TauObject) Get(n string) (Object, bool) {
	o, ok := to[n]
	return o, ok
}

func (to TauObject) Set(n string, o Object) Object {
	to[n] = o
	return o
}

func (to TauObject) Module() (o TauObject) {
	for k, v := range to {
		if isExported(k) {
			o[k] = v
		}
	}
	return
}

func isExported(n string) bool {
	r, _ := utf8.DecodeRuneInString(n)
	return unicode.IsUpper(r)
}
