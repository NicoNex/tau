package obj

type wrappedGetSetter struct {
	outer GetSetter
	n     string
}

func NewWrappedGetSetter(gs GetSetter, n string) Object {
	return &wrappedGetSetter{gs, n}
}

func (w wrappedGetSetter) Type() Type {
	return w.outer.Type()
}

func (w wrappedGetSetter) String() string {
	return w.outer.String()
}

func (w wrappedGetSetter) Set(o Object) Object {
	return w.outer.Set(o)
}

// func (w wrappedGetSetter) Get() (Object, bool) {
// 	return w.outer.Get(w.n)
// }

func (w wrappedGetSetter) Object() Object {
	return w.outer.Object()
}
