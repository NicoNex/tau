package obj

import (
	"fmt"
	"strings"
)

type Map map[KeyHash]MapPair

type MapPair struct {
	Key   Object
	Value Object
}

func NewMap() Map {
	return Map(make(map[KeyHash]MapPair))
}

func (m Map) Set(k KeyHash, v MapPair) {
	m[k] = v
}

func (m Map) Get(k KeyHash) MapPair {
	if v, ok := m[k]; ok {
		return v
	}
	return MapPair{Key: NullObj, Value: NullObj}
}

func (m Map) Type() Type {
	return MapType
}

func (m Map) String() string {
	var buf strings.Builder
	var i = 1

	buf.WriteString("{")
	for _, v := range m {
		buf.WriteString(fmt.Sprintf("%v: %v", v.Key, v.Value))

		if i < len(m) {
			buf.WriteString(", ")
		}
		i += 1
	}
	buf.WriteString("}")

	return buf.String()
}
