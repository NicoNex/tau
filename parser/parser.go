package parser

import (
	"errors"
	"strconv"

	"github.com/NicoNex/calc/ast"
	"github.com/NicoNex/calc/utils"
)

// Type used to abstract the constructor functions of the operators.
type newOp func(ast.Node, ast.Node) ast.Node

var opFuncs = map[string]newOp{
	"+": ast.NewPlus,
	"-": ast.NewMinus,
	"*": ast.NewTimes,
	"/": ast.NewDivide,
	"^": ast.NewPower,
}

var precedence = map[string]int{
	"+": 0,
	"-": 0,
	"*": 1,
	"/": 1,
	"^": 2,
	"=": 3,
}

var input string

// Converts a string operand to a float64 and returns it.
func parseOperand(o string) (float64, error) {
	return strconv.ParseFloat(o, 64)
}

// Returns the AST generated from the operators stack and operands queue.
func genAst(expr []item) (ast.Node, error) {
	var output = utils.NewStack()

	for i, itm := range expr {
		switch itm.typ {
		case itemOperand:
			val, err := parseOperand(itm.val)
			if err != nil {
				return nil, NewSyntaxError("invalid operand", input, itm.pos)
			}
			output.Push(ast.NewConst(val))

		case itemOperator:
			rnode := output.Pop()
			if rnode == nil {
				return nil, NewSyntaxError("invalid operator", input, itm.pos)
			}
			lnode := output.Pop()
			if lnode == nil {
				return nil, NewSyntaxError("invalid operator", input, itm.pos)
			}
			if fn, ok := opFuncs[itm.val]; ok {
				output.Push(fn(lnode.(ast.Node), rnode.(ast.Node)))
			} else {
				return nil, NewSyntaxError("invalid operator", input, itm.pos)
			}

		case itemVariable:
			output.Push(ast.NewVariable(itm.val))

		case itemAssign:
			output.Pop()
			v := output.Pop()
			if v == nil {
				return nil, NewSyntaxError("invalid statement", input, itm.pos)
			}
			if i > 0 {
				tmp := append(expr[1:i], expr[i+1:]...)
				right, err := genAst(tmp)
				if err != nil {
					return nil, err
				}
				va, ok := v.(ast.Variable)
				if !ok {
					return nil, NewSyntaxError("invalid assignment", input, itm.pos)
				}
				return ast.NewAssign(va, right), nil
			}
			return nil, NewSyntaxError("invalid statement", input, itm.pos)

		default:
			return nil, NewSyntaxError("invalid syntax", input, itm.pos)
		}
	}

	if ret := output.Pop(); ret != nil {
		return ret.(ast.Node), nil
	}
	return nil, errors.New("syntax error")
}

// Returns true if a has precedence over b.
func hasPrecendence(a, b item) bool {
	return precedence[a.val] > precedence[b.val]
}

func toPostfix(items chan item) []item {
	var ret []item
	var stack = utils.NewStack()

	for i := range items {
		switch i.typ {
		case itemOperand, itemVariable, itemError:
			ret = append(ret, i)

		case itemOperator, itemAssign:
			for o := stack.Peek(); o != nil; o = stack.Peek() {
				if !hasPrecendence(o.(item), i) {
					break
				}
				ret = append(ret, o.(item))
				stack.Pop()
			}
			stack.Push(i)

		case itemBracket:
			switch i.val {
			case "(":
				stack.Push(i)
			case ")":
				for o := stack.Pop(); o != nil; o = stack.Pop() {
					if tmp := o.(item); tmp.val == "(" {
						break
					}
					ret = append(ret, o.(item))
				}
			}
		}
	}

	for o := stack.Pop(); o != nil; o = stack.Pop() {
		ret = append(ret, o.(item))
	}
	return ret
}

// Evaluates the types from the lexer and returns the AST.
func Parse(a string) (ast.Node, error) {
	_, items := lex(a)
	input = a
	return genAst(toPostfix(items))
}
