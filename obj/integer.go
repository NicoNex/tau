package obj

import "strconv"

type Integer struct {
	i int64
	*Env
}

func NewInteger(i int64) Object {
	return &Integer{i, NewEnv()}
}

func (i Integer) String() string {
	return strconv.FormatInt(i.i, 10)
}

func (i Integer) Val() int64 {
	return i.i
}

func (i Integer) Type() Type {
	return INT
}
