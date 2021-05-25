package obj

type Builtin func(arg ...Object) Object

func (b Builtin) Type() Type {
	return BUILTIN
}

func (b Builtin) Get(n string) (Object, bool) {
	return nil, false
}

func (b Builtin) Set(n string, o Object) Object {
	return nil
}

func (b Builtin) String() string {
	return "<builtin function>"
}
