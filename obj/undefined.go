package obj

type Undefined struct {
	s setter
	n string
}

func NewUndefined(s setter, n string) Object {
	return &Undefined{s, n}
}

func (u Undefined) Type() Type {
	return NullType
}

func (u Undefined) String() string {
	return NullObj.String()
}

func (u Undefined) Set(o Object) Object {
	return u.s.Set(u.n, o)
}

func (u Undefined) Object() Object {
	return NullObj
}
