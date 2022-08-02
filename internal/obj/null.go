package obj

type Null struct{}

func NewNull() Object {
	return new(Null)
}

func (n Null) String() string {
	return "null"
}

func (n Null) Type() Type {
	return NullType
}

func (n Null) Val() interface{} {
	return nil
}
