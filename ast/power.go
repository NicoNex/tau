package ast

import (
	"fmt"
	"math"
)

type Power struct {
	l Node
	r Node
}

func NewPower(l, r Node) Node {
	return Power{
		l,
		r,
	}
}

func (p Power) Eval() float64 {
	return math.Pow(p.l.Eval(), p.r.Eval())
}

func (p Power) String() string {
	return fmt.Sprintf("(%v^%v)", p.l, p.r)
}
