package obj

type Null struct{}

func (n Null) String() string {
	return "null"
}

func (n Null) Type() Type {
	return NULL
}
