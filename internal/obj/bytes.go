package obj

import (
	"strconv"
	"strings"
)

type Bytes []byte

func NewBytes(b []byte) Object {
	return Bytes(b)
}

func (b Bytes) Type() Type {
	return BytesType
}

func (bytes Bytes) String() string {
	var buf strings.Builder

	buf.WriteByte('[')
	for i, b := range bytes {
		buf.WriteString(strconv.Itoa(int(b)))
		if i < len(bytes) - 1 {
			buf.WriteString(", ")		
		}
	}
	buf.WriteByte(']')

	return buf.String()
}

func (b Bytes) Val() []byte {
	return []byte(b) 
}
