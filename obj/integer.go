package obj

import "strconv"

type Integer int64

func NewInteger(i int64) Object {
	var ret = Integer(i)
	return &ret
}

func ObjectToInt(i Object) (int64, bool) {
	a := i.String()
	if b, ok := strconv.ParseFloat(string(a), 64); ok == nil {
		return int64(b), true
	}
	return -1, false
}

func (i Integer) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i Integer) Type() Type {
	return INT
}

func (i Integer) Val() int64 {
	return int64(i)
}
