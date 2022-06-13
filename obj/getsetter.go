package obj

type GetSetterImpl struct {
	GetFunc func() (Object, bool)
	SetFunc func(Object) Object
}

func (g GetSetterImpl) Object() Object {
	if g.GetFunc == nil {
		return NullObj
	}
	if o, ok := g.GetFunc(); ok {
		return o
	}
	return NullObj
}

func (g GetSetterImpl) Set(o Object) Object {
	return g.SetFunc(o)
}

func (g GetSetterImpl) Type() Type {
	if g.GetFunc == nil {
		return NullType
	}
	if o, ok := g.GetFunc(); ok {
		return o.Type()
	}
	return NullType
}

func (g GetSetterImpl) String() string {
	if g.GetFunc == nil {
		return NullObj.String()
	}
	if o, ok := g.GetFunc(); ok {
		return o.String()
	}
	return NullObj.String()
}
