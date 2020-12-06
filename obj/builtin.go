package obj

type Builtin func(arg ...Object) Object

func (b Builtin) Type() Type {
	return BUILTIN
}

func (b Builtin) String() string {
	return "<builtin function>"
}
