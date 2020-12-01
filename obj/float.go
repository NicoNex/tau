package obj

import "strconv"

type Float float64

func NewFloat(f float64) Object {
	return &Float(f)
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
