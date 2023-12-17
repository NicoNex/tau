package ast

func NewPlusPlus(r Node, pos int) Node {
	return Assign{l: r, r: Plus{l: r, r: Integer(1), pos: pos}, pos: pos}
}
