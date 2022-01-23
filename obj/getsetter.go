package obj

type getsetter struct {
	s MapGetSetter
	n string
}

func NewGetSetter(s MapGetSetter, n string) Object {
	return &getsetter{s, n}
}

func (g getsetter) Type() Type {
	if o, ok := g.s.Get(g.n); ok {
		return o.Type()
	}
	return NullType
}

func (g getsetter) String() string {
	if o, ok := g.s.Get(g.n); ok {
		return o.String()
	}
	return NullObj.String()
}

func (g getsetter) Get() (Object, bool) {
	return g.s.Get(g.n)
}

func (g getsetter) Set(o Object) Object {
	return g.s.Set(g.n, o)
}

func (g getsetter) Object() Object {
	if o, ok := g.s.Get(g.n); ok {
		return o
	}
	return NullObj
}
