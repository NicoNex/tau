package parser

import (
	"fmt"
	"strconv"

	"tau/ast"
	"tau/item"
	"tau/lexer"
)

type Parser struct {
	cur           item.Item
	peek          item.Item
	items         chan item.Item
	errs          []string
	prefixParsers map[item.Type]parsePrefixFn
	infixParsers  map[item.Type]parseInfixFn
}

type (
	parsePrefixFn func() ast.Node
	parseInfixFn  func(ast.Node) ast.Node
)

// Operators' precedence classes.
const (
	LOWEST int = iota
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
	INDEX
)

// Links each operator to its precedence class.
var precedences = map[item.Type]int{
	item.EQ:       EQUALS,
	item.NOT_EQ:   EQUALS,
	item.LT:       LESSGREATER,
	item.GT:       LESSGREATER,
	item.LT_EQ:    LESSGREATER,
	item.GT_EQ:    LESSGREATER,
	item.PLUS:     SUM,
	item.MINUS:    SUM,
	item.SLASH:    PRODUCT,
	item.ASTERISK: PRODUCT,
	item.POWER:    PRODUCT,
	item.LPAREN:   CALL,
	item.LBRACKET: INDEX,
}

func newParser(items chan item.Item) *Parser {
	p := &Parser{
		cur:           <-items,
		peek:          <-items,
		items:         items,
		prefixParsers: make(map[item.Type]parsePrefixFn),
		infixParsers:  make(map[item.Type]parseInfixFn),
	}
	// p.registerPrefix(item.IDENT, p.parseIdentifier)
	p.registerPrefix(item.INT, p.parseInteger)
	p.registerPrefix(item.FLOAT, p.parseFloat)
	// p.registerPrefix(item.STRING, p.parseStringLiteral)
	p.registerPrefix(item.MINUS, p.parsePrefixMinus)
	p.registerPrefix(item.BANG, p.parseBang)
	p.registerPrefix(item.TRUE, p.parseBoolean)
	p.registerPrefix(item.FALSE, p.parseBoolean)
	// p.registerPrefix(item.LPAREN, p.parseGroupedExpression)
	// p.registerPrefix(item.IF, p.parseIfExpression)
	// p.registerPrefix(item.FUNCTION, p.parseFunctionLiteral)
	// p.registerPrefix(item.LBRACKET, p.parseArrayLiteral)

	p.registerInfix(item.EQ, p.parseEquals)
	// p.registerInfix(item.NOT_EQ, p.parseInfixExpression)
	// p.registerInfix(item.LT, p.parseInfixExpression)
	// p.registerInfix(item.GT, p.parseInfixExpression)
	// p.registerInfix(item.LT_EQ, p.parseInfixExpression)
	// p.registerInfix(item.GT_EQ, p.parseInfixExpression)
	p.registerInfix(item.PLUS, p.parseInfixExpression)
	p.registerInfix(item.MINUS, p.parseInfixExpression)
	p.registerInfix(item.SLASH, p.parseInfixExpression)
	p.registerInfix(item.ASTERISK, p.parseInfixExpression)
	// p.registerInfix(item.POWER, p.parseInfixExpression)
	// p.registerInfix(item.LPAREN, p.parseCallExpression)
	// p.registerInfix(item.LBRACKET, p.parseIndexExpression)
	return p
}

func (p *Parser) next() {
	p.cur = p.peek
	p.peek = <-p.items
}

func (p *Parser) errors() []string {
	return p.errs
}

func (p *Parser) parse() ast.Node {
	var block = ast.NewBlock()

	for !p.cur.Is(item.EOF) {
		if s := p.parseStatement(); s != nil {
			block.Add(s)
		}
		p.next()
	}
	return block
}

func (p *Parser) parseStatement() ast.Node {
	if p.cur.Is(item.RETURN) {
		return p.parseReturn()
	}
	return p.parseExpr(LOWEST)
}

func (p *Parser) parseReturn() ast.Node {
	p.next()
	var ret = ast.NewReturn(p.parseExpr(LOWEST))

	if p.peek.Is(item.SEMICOLON) {
		p.next()
	}
	return ret
}

func (p *Parser) parseExpr(precedence int) ast.Node {
	if prefixFn, ok := p.prefixParsers[p.cur.Typ]; ok {
		leftExp := prefixFn()

		for !p.peek.Is(item.SEMICOLON) && precedence < p.peekPrecedence() {
			if infixFn, ok := p.infixParsers[p.peek.Typ]; ok {
				p.next()
				leftExp = infixFn(leftExp)
			} else {
				break
			}
		}
		return leftExp
	}
	p.noParsePrefixFnError(p.cur.Typ)
	return nil
}

// Returns an integer node.
func (p *Parser) parseInteger() ast.Node {
	i, err := strconv.ParseInt(p.cur.Val, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("unable to parse %q as integer", p.cur.Val)
		p.errs = append(p.errs, msg)
		return nil
	}
	return ast.NewInteger(i)
}

// Returns a float node.
func (p *Parser) parseFloat() ast.Node {
	f, err := strconv.ParseFloat(p.cur.Val, 64)
	if err != nil {
		msg := fmt.Sprintf("unable to parse %q as float", p.cur.Val)
		p.errs = append(p.errs, msg)
		return nil
	}
	return ast.NewFloat(f)
}

// Returns an identifier node.
// func (p *Parser) parseIdentifier() ast.Identifier {
// 	return ast.NewIdentifier(p.cur.Val)
// }

// Returns a boolean node.
func (p *Parser) parseBoolean() ast.Node {
	return ast.NewBoolean(p.cur.Is(item.TRUE))
}

// Returns a node of type PrefixMinus.
func (p *Parser) parsePrefixMinus() ast.Node {
	p.next()
	return ast.NewPrefixMinus(p.parseExpr(PREFIX))
}

// Returns a node of type Bang.
func (p *Parser) parseBang() ast.Node {
	p.next()
	return ast.NewBang(p.parseExpr(PREFIX))
}

// Returns a node of type ast.Equals.
func (p *Parser) parseEquals(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewEquals(left, p.parseExpr(prec))
}

// Returns the expression obtained by parsin an infix expression.
func (p *Parser) parseInfixExpression(left ast.Node) ast.Node {
	var prec = p.precedence()
	var typ = p.cur.Typ

	p.next()
	switch typ {
	case item.PLUS:
		return ast.NewPlus(left, p.parseExpr(prec))

	case item.MINUS:
		return ast.NewMinus(left, p.parseExpr(prec))

	case item.ASTERISK:
		return ast.NewTimes(left, p.parseExpr(prec))

	case item.SLASH:
		return ast.NewDivide(left, p.parseExpr(prec))

	default:
		msg := fmt.Sprintf("cannot parse %v infix operator", p.cur)
		p.errs = append(p.errs, msg)
		return nil
	}
}

// Returns the precedence value of the type of the peek item.
func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peek.Typ]; ok {
		return prec
	}
	return LOWEST
}

// Returns the precedence value of the type of the current item.
func (p *Parser) precedence() int {
	if prec, ok := precedences[p.cur.Typ]; ok {
		return prec
	}
	return LOWEST
}

// Adds fn to the prefix parsers table with key 'typ'.
func (p *Parser) registerPrefix(typ item.Type, fn parsePrefixFn) {
	p.prefixParsers[typ] = fn
}

// Adds fn to the infix parsers table with key 'typ'.
func (p *Parser) registerInfix(typ item.Type, fn parseInfixFn) {
	p.infixParsers[typ] = fn
}

func (p *Parser) noParsePrefixFnError(t item.Type) {
	msg := fmt.Sprintf("no parse prefix function for '%s' found", t)
	p.errs = append(p.errs, msg)
}

func Parse(input string) (prog ast.Node, errs []string) {
	items := lexer.Lex(input)
	p := newParser(items)
	return p.parse(), p.errors()
}
