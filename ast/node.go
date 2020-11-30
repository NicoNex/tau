package ast

type Node interface {
	Eval() float64
	String() string
}
