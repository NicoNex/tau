package obj

type String string

func NewString(s string) Object {
	var ret = String(s)
	return &ret
}

func (s String) Type() Type {
	return STRING
}

func (s String) String() string {
	return string(s)
}

func (s String) Val() string {
	return string(s)
}
