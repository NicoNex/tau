package obj

type Null struct {
	*Env
}

func NewNull() Object {
	return &Null{NewEnv()}
}

func (n Null) String() string {
	return "null"
}

func (n Null) Type() Type {
	return NULL
}
