package obj

type String struct {
	s string
	*Env
}

func NewString(s string) Object {
	return &String{s, NewEnv()}
}

func (s String) Type() Type {
	return STRING
}

func (s String) Len() int {
	return len(s.s)
}

func (s String) String() string {
	return s.s
}

func (s String) Val() string {
	return s.s
}
