package obj

import "fmt"

type Return struct {
	v Object
}

func NewReturn(o Object) Object {
	return &Return{o}
}

func (r Return) String() string {
	return fmt.Sprintf("return %v;", r.v)
}

func (r Return) Type() Type {
	return ReturnType
}

func (r Return) Val() Object {
	return r.v
}
