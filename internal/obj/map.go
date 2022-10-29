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
		var key, val string

		if s, ok := v.Key.(String); ok {
			key = s.Quoted()
		} else {
			key = v.Key.String()
		}

		if s, ok := v.Value.(String); ok {
			val = s.Quoted()
		} else {
			val = v.Value.String()
		}

		buf.WriteString(fmt.Sprintf("%s: %s", key, val))

		if i < len(m) {
			buf.WriteString(", ")
		}
		i += 1
	}
	buf.WriteString("}")

	return buf.String()
}
