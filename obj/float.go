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

func ObjectToFloat(i Object) (float64, bool) {
	a := i.String()
	if b, ok := strconv.ParseFloat(string(a), 64); ok == nil {
		return b, true
	}
	return -1, false
}

func (f Float) Type() Type {
	return FLOAT
}

func (f Float) Val() float64 {
	return float64(f)
}
