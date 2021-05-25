package obj

import "strconv"

type Boolean struct {
	b bool
	*Env
}

func NewBoolean(b bool) Object {
	return &Boolean{
		b: b,
		Env: NewEnv(),
	}
}

func (b Boolean) String() string {
	return strconv.FormatBool(b.b)
}

func (b Boolean) Type() Type {
	return BOOL
}

func (b Boolean) Val() bool {
	return b.b
}
