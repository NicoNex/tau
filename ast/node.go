package ast

import "tau/obj"

type Node interface {
	Eval() obj.Object
	String() string
}
