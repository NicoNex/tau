package obj

type Setter interface {
	Set(string, Object) Object
}

type Undefined struct {
	s Setter
	n string
}

func NewUndefined(s Setter, n string) Object {
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
