package ast

import "fmt"

type Assign struct {
	l Node
	r Node
}

var vtable map[string]float64

func NewAssign(l, r Node) Node {
	return Assign{
		l,
		r,
	}
}

func (a Assign) Eval() float64 {
	var v = a.r.Eval()

	vtable[a.l.String()] = v
	return v
}

func (a Assign) String() string {
	return fmt.Sprintf("(%v=%v)", a.l, a.r)
}

func init() {
	vtable = make(map[string]float64)
}
