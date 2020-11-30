package ast

import "strconv"

type Const struct {
	v float64
}

func NewConst(v float64) Node {
	return Const{v}
}

func (c Const) Eval() float64 {
	return c.v
}

func (c Const) String() string {
	return strconv.FormatFloat(c.v, 'f', -1, 64)
}
