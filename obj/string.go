package obj

type String struct {
	Value string
}

func (s *String) Type() Type {
	return STRING
}

func (s *String) String() string {
	return string(s)
}

func (s String) Val() string {
	return string(s)
}
