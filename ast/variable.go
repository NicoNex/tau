package ast

type Variable struct {
	n string
}

func NewVariable(n string) Node {
	return Variable{
		n,
	}
}

func (v Variable) Eval() float64 {
	return vtable[v.n]
}

func (v Variable) String() string {
	return v.n
}
