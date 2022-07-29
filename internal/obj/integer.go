package obj

import "strconv"

type Integer int64

func NewInteger(i int64) Object {
	var ret = Integer(i)
	return &ret
}

func (i Integer) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i Integer) Type() Type {
	return IntType
}

func (i Integer) Val() int64 {
	return int64(i)
}

func (i Integer) KeyHash() KeyHash {
	return KeyHash{Type: IntType, Value: uint64(i)}
}
