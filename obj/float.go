package obj

import "strconv"

type Float struct {
	f float64
	*Env
}

func NewFloat(f float64) Object {
	return &Float{
		f: f,
		Env: NewEnv(),
	}
}

func (f Float) String() string {
	return strconv.FormatFloat(f.f, 'f', -1, 64)
}

func (f Float) Type() Type {
	return FLOAT
}

func (f Float) Val() float64 {
	return f.f
}
