package obj

import "strconv"

type Float float64

func NewFloat(f float64) Object {
	var ret = Float(f)
	return &ret
}

func (f Float) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f Float) Type() Type {
	return FLOAT
}

func (f Float) Val() float64 {
	return float64(f)
}

func (f Float) KeyHash() KeyHash {
	return KeyHash{Type: FLOAT, Value: uint64(f)}
}
