package parser

import (
	"fmt"
	"strconv"

	"github.com/NicoNex/tau/internal/ast"
	"github.com/NicoNex/tau/internal/item"
	"github.com/NicoNex/tau/internal/lexer"
)

type Parser struct {
	cur           item.Item
	peek          item.Item
	items         chan item.Item
	errs          []string
	nestedLoops   uint
	prefixParsers map[item.Type]parsePrefixFn
	infixParsers  map[item.Type]parseInfixFn
}

type (
	parsePrefixFn func() ast.Node
	parseInfixFn  func(ast.Node) ast.Node
)

// Operators' precedence classes.
const (
	Lowest int = iota
	Assignment
	LogicalOr
	LogicalAnd
	BitwiseOr
	BitwiseXor
	BitwiseAnd
	Equality
	Relational
	Shift
	Additive
	Multiplicative
	Prefix
	Call
	Index
	Dot
)

// Links each operator to its precedence class.
var precedences = map[item.Type]int{
	item.Assign:         Assignment,
	item.PlusAssign:     Assignment,
	item.MinusAssign:    Assignment,
	item.SlashAssign:    Assignment,
	item.AsteriskAssign: Assignment,
	item.ModulusAssign:  Assignment,
	item.BwAndAssign:    Assignment,
	item.BwOrAssign:     Assignment,
	item.BwXorAssign:    Assignment,
	item.LShiftAssign:   Assignment,
	item.RShiftAssign:   Assignment,
	item.Or:             LogicalOr,
	item.And:            LogicalAnd,
	item.Equals:         Equality,
	item.NotEquals:      Equality,
	item.In:             Relational,
	item.LT:             Relational,
	item.GT:             Relational,
	item.LTEQ:           Relational,
	item.GTEQ:           Relational,
	item.Plus:           Additive,
	item.Minus:          Additive,
	item.Modulus:        Multiplicative,
	item.Slash:          Multiplicative,
	item.Asterisk:       Multiplicative,
	item.PlusPlus:       Prefix,
	item.MinusMinus:     Prefix,
	item.BwAnd:          BitwiseAnd,
	item.BwOr:           BitwiseOr,
	item.BwXor:          BitwiseOr,
	item.LShift:         Shift,
	item.RShift:         Shift,
	item.LParen:         Call,
	item.LBracket:       Index,
	item.Dot:            Dot,
}

func newParser(items chan item.Item) *Parser {
	p := &Parser{
		cur:           <-items,
		peek:          <-items,
		items:         items,
		prefixParsers: make(map[item.Type]parsePrefixFn),
		infixParsers:  make(map[item.Type]parseInfixFn),
	}
	p.registerPrefix(item.Ident, p.parseIdentifier)
	p.registerPrefix(item.Int, p.parseInteger)
	p.registerPrefix(item.Float, p.parseFloat)
	p.registerPrefix(item.String, p.parseString)
	p.registerPrefix(item.RawString, p.parseRawString)
	p.registerPrefix(item.Minus, p.parsePrefixMinus)
	p.registerPrefix(item.Bang, p.parseBang)
	p.registerPrefix(item.True, p.parseBoolean)
	p.registerPrefix(item.False, p.parseBoolean)
	p.registerPrefix(item.LParen, p.parseGroupedExpr)
	p.registerPrefix(item.If, p.parseIfExpr)
	p.registerPrefix(item.Function, p.parseFunction)
	p.registerPrefix(item.LBracket, p.parseList)
	p.registerPrefix(item.PlusPlus, p.parsePlusPlus)
	p.registerPrefix(item.MinusMinus, p.parseMinusMinus)
	p.registerPrefix(item.For, p.parseFor)
	p.registerPrefix(item.LBrace, p.parseMap)
	p.registerPrefix(item.Null, p.parseNull)
	p.registerPrefix(item.BwNot, p.parseBwNot)
	p.registerPrefix(item.Continue, p.parseContinue)
	p.registerPrefix(item.Break, p.parseBreak)
	p.registerPrefix(item.Import, p.parseImport)
	p.registerPrefix(item.Error, p.parseError)

	p.registerInfix(item.Equals, p.parseEquals)
	p.registerInfix(item.NotEquals, p.parseNotEquals)
	p.registerInfix(item.LT, p.parseLess)
	p.registerInfix(item.GT, p.parseGreater)
	p.registerInfix(item.LTEQ, p.parseLessEq)
	p.registerInfix(item.GTEQ, p.parseGreaterEq)
	p.registerInfix(item.And, p.parseAnd)
	p.registerInfix(item.Or, p.parseOr)
	p.registerInfix(item.In, p.parseIn)
	p.registerInfix(item.Plus, p.parsePlus)
	p.registerInfix(item.Minus, p.parseMinus)
	p.registerInfix(item.Slash, p.parseSlash)
	p.registerInfix(item.Asterisk, p.parseAsterisk)
	p.registerInfix(item.Modulus, p.parseModulus)
	p.registerInfix(item.BwAnd, p.parseBwAnd)
	p.registerInfix(item.BwOr, p.parseBwOr)
	p.registerInfix(item.BwXor, p.parseBwXor)
	p.registerInfix(item.LShift, p.parseLShift)
	p.registerInfix(item.RShift, p.parseRShift)
	p.registerInfix(item.Assign, p.parseAssign)
	p.registerInfix(item.PlusAssign, p.parsePlusAssign)
	p.registerInfix(item.MinusAssign, p.parseMinusAssign)
	p.registerInfix(item.SlashAssign, p.parseSlashAssign)
	p.registerInfix(item.AsteriskAssign, p.parseAsteriskAssign)
	p.registerInfix(item.ModulusAssign, p.parseModulusAssign)
	p.registerInfix(item.BwAndAssign, p.parseBwAndAssign)
	p.registerInfix(item.BwOrAssign, p.parseBwOrAssign)
	p.registerInfix(item.BwXorAssign, p.parseBwXorAssign)
	p.registerInfix(item.LShiftAssign, p.parseLShiftAssign)
	p.registerInfix(item.RShiftAssign, p.parseRShiftAssign)
	p.registerInfix(item.LParen, p.parseCall)
	p.registerInfix(item.LBracket, p.parseIndex)
	p.registerInfix(item.Dot, p.parseDot)

	return p
}

func (p *Parser) enterLoop() {
	p.nestedLoops += 1
}

func (p *Parser) exitLoop() {
	p.nestedLoops -= 1
}

func (p *Parser) isInsideLoop() bool {
	return p.nestedLoops > 0
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
	if p.cur.Is(item.Return) {
		return p.parseReturn()
	}
	return p.parseExpr(Lowest)
}

func (p *Parser) parseReturn() ast.Node {
	p.next()
	var ret = ast.NewReturn(p.parseExpr(Lowest))

	if p.peek.Is(item.Semicolon) {
		p.next()
	}
	return ret
}

func (p *Parser) parseExpr(precedence int) ast.Node {
	if prefixFn, ok := p.prefixParsers[p.cur.Typ]; ok {
		leftExp := prefixFn()

		for !p.peek.Is(item.Semicolon) && precedence < p.peekPrecedence() {
			if infixFn, ok := p.infixParsers[p.peek.Typ]; ok {
				p.next()
				leftExp = infixFn(leftExp)
			} else {
				break
			}
		}

		if p.peek.Is(item.Semicolon) {
			p.next()
		}
		return leftExp
	}
	p.noParsePrefixFnError(p.cur.Typ)
	return nil
}

// Returns the node representing an expression enclosed in parentheses.
func (p *Parser) parseGroupedExpr() ast.Node {
	p.next()
	exp := p.parseExpr(Lowest)
	if !p.expectPeek(item.RParen) {
		return nil
	}
	return exp
}

// Returns the node representing a series of expressions enclosed in curly braces.
func (p *Parser) parseBlock() ast.Node {
	var block ast.Block
	p.next()

	for !p.cur.Is(item.RBrace) && !p.cur.Is(item.EOF) {
		if s := p.parseStatement(); s != nil {
			block.Add(s)
		}
		p.next()
	}
	return block
}

func (p *Parser) parseIfExpr() ast.Node {
	p.next()
	cond := p.parseExpr(Lowest)

	if !p.expectPeek(item.LBrace) {
		return nil
	}

	body := p.parseBlock()

	var alt ast.Node
	if p.peek.Is(item.Else) {
		p.next()

		if p.peek.Is(item.If) {
			p.next()
			alt = p.parseIfExpr()
		} else {
			if !p.expectPeek(item.LBrace) {
				return nil
			}
			alt = p.parseBlock()
		}
	}

	return ast.NewIfExpr(cond, body, alt)
}

func (p *Parser) parseList() ast.Node {
	nodes := p.parseNodeList(item.RBracket)
	return ast.NewList(nodes...)
}

func (p *Parser) parseMap() ast.Node {
	couples := p.parseNodePairs(item.RBrace)
	return ast.NewMap(couples...)
}

func (p *Parser) parseImport() ast.Node {
	if !p.expectPeek(item.LParen) {
		return nil
	}

	args := p.parseNodeList(item.RParen)

	if l := len(args); l != 1 {
		p.errs = append(p.errs, fmt.Sprintf("import: expected exactly 1 argument but %d provided", l))
		return nil
	}

	return ast.NewImport(args[0], Parse)
}

func (p *Parser) parseFunction() ast.Node {
	if !p.expectPeek(item.LParen) {
		return nil
	}

	params := p.parseFunctionParams()
	if !p.expectPeek(item.LBrace) {
		return nil
	}

	body := p.parseBlock()
	return ast.NewFunction(params, body)
}

func (p *Parser) parseFunctionParams() []ast.Identifier {
	var ret []ast.Identifier

	if p.peek.Is(item.RParen) {
		p.next()
		return ret
	}

	p.next()
	ret = append(ret, ast.Identifier(p.cur.Val))

	for p.peek.Is(item.Comma) {
		p.next()
		p.next()
		ret = append(ret, ast.Identifier(p.cur.Val))
	}

	if !p.expectPeek(item.RParen) {
		return nil
	}
	return ret
}

// Returns an identifier node.
func (p *Parser) parseIdentifier() ast.Node {
	return ast.NewIdentifier(p.cur.Val)
}

func (p *Parser) parseNull() ast.Node {
	return ast.NewNull()
}

func (p *Parser) parseContinue() ast.Node {
	if !p.isInsideLoop() {
		p.errs = append(p.errs, `continue statement not inside "for" block`)
		return nil
	}
	return ast.NewContinue()
}

func (p *Parser) parseBreak() ast.Node {
	if !p.isInsideLoop() {
		p.errs = append(p.errs, `break statement not inside "for" block`)
		return nil
	}
	return ast.NewBreak()
}

func (p *Parser) parseError() ast.Node {
	p.errs = append(p.errs, p.cur.Val)
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

func (p *Parser) parseString() ast.Node {
	s, err := ast.NewString(p.cur.Val)
	if err != nil {
		p.errs = append(p.errs, err.Error())
		return nil
	}
	return s
}

func (p *Parser) parseRawString() ast.Node {
	return ast.NewRawString(p.cur.Val)
}

// Returns a boolean node.
func (p *Parser) parseBoolean() ast.Node {
	return ast.NewBoolean(p.cur.Is(item.True))
}

// Returns a node of type PrefixMinus.
func (p *Parser) parsePrefixMinus() ast.Node {
	p.next()
	return ast.NewPrefixMinus(p.parseExpr(Prefix))
}

func (p *Parser) parsePlusPlus() ast.Node {
	p.next()
	return ast.NewPlusPlus(p.parseExpr(Prefix))
}

func (p *Parser) parseMinusMinus() ast.Node {
	p.next()
	return ast.NewMinusMinus(p.parseExpr(Prefix))
}

func (p *Parser) parseFor() ast.Node {
	var arg []ast.Node
	p.enterLoop()
	defer p.exitLoop()

	p.next()
	if p.cur.Is(item.LBrace) {
		return ast.NewFor(ast.NewBoolean(true), p.parseBlock(), nil, nil)
	}

	for !p.cur.Is(item.LBrace) && !p.cur.Is(item.EOF) {
		arg = append(arg, p.parseExpr(Lowest))
		p.next()
	}

	switch l := len(arg); l {
	case 1:
		return ast.NewFor(arg[0], p.parseBlock(), nil, nil)

	case 3:
		return ast.NewFor(arg[1], p.parseBlock(), arg[0], arg[2])

	default:
		msg := fmt.Sprintf("wrong number of expressions, expected 1 or 3 but got %d", l)
		p.errs = append(p.errs, msg)
		return nil
	}
}

// Returns a node of type Bang.
func (p *Parser) parseBang() ast.Node {
	p.next()
	return ast.NewBang(p.parseExpr(Prefix))
}

func (p *Parser) parsePlus(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewPlus(left, p.parseExpr(prec))
}

func (p *Parser) parseMinus(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewMinus(left, p.parseExpr(prec))
}

func (p *Parser) parseAsterisk(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewTimes(left, p.parseExpr(prec))
}

func (p *Parser) parseSlash(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewDivide(left, p.parseExpr(prec))
}

func (p *Parser) parseModulus(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewMod(left, p.parseExpr(prec))
}

func (p *Parser) parseBwAnd(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewBitwiseAnd(left, p.parseExpr(prec))
}

func (p *Parser) parseBwNot() ast.Node {
	p.next()
	return ast.NewBitwiseNot(p.parseExpr(Prefix))
}

func (p *Parser) parseBwOr(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewBitwiseOr(left, p.parseExpr(prec))
}

func (p *Parser) parseBwXor(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewBitwiseXor(left, p.parseExpr(prec))
}

func (p *Parser) parseLShift(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewBitwiseLeftShift(left, p.parseExpr(prec))
}

func (p *Parser) parseRShift(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewBitwiseRightShift(left, p.parseExpr(prec))
}

// Returns a node of type ast.Equals.
func (p *Parser) parseEquals(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewEquals(left, p.parseExpr(prec))
}

// Returns a node of type ast.Equals.
func (p *Parser) parseNotEquals(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewNotEquals(left, p.parseExpr(prec))
}

func (p *Parser) parseLess(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewLess(left, p.parseExpr(prec))
}

func (p *Parser) parseGreater(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewGreater(left, p.parseExpr(prec))
}

func (p *Parser) parseLessEq(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewLessEq(left, p.parseExpr(prec))
}

func (p *Parser) parseGreaterEq(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewGreaterEq(left, p.parseExpr(prec))
}

func (p *Parser) parseAnd(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewAnd(left, p.parseExpr(prec))
}

func (p *Parser) parseOr(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewOr(left, p.parseExpr(prec))
}

func (p *Parser) parseIn(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewIn(left, p.parseExpr(prec))
}

func (p *Parser) parseAssign(left ast.Node) ast.Node {
	p.next()
	right := p.parseExpr(Lowest)

	i, leftIsIdentifier := left.(ast.Identifier)
	fn, rightIsFunction := right.(ast.Function)

	if leftIsIdentifier && rightIsFunction {
		fn.Name = i.String()
	}

	return ast.NewAssign(left, right)
}

func (p *Parser) parsePlusAssign(left ast.Node) ast.Node {
	p.next()
	return ast.NewPlusAssign(left, p.parseExpr(Lowest))
}

func (p *Parser) parseMinusAssign(left ast.Node) ast.Node {
	p.next()
	return ast.NewMinusAssign(left, p.parseExpr(Lowest))
}

func (p *Parser) parseSlashAssign(left ast.Node) ast.Node {
	p.next()
	return ast.NewDivideAssign(left, p.parseExpr(Lowest))
}

func (p *Parser) parseAsteriskAssign(left ast.Node) ast.Node {
	p.next()
	return ast.NewTimesAssign(left, p.parseExpr(Lowest))
}

func (p *Parser) parseModulusAssign(left ast.Node) ast.Node {
	p.next()
	return ast.NewModAssign(left, p.parseExpr(Lowest))
}

func (p *Parser) parseBwAndAssign(left ast.Node) ast.Node {
	p.next()
	return ast.NewBitwiseAndAssign(left, p.parseExpr(Lowest))
}

func (p *Parser) parseBwOrAssign(left ast.Node) ast.Node {
	p.next()
	return ast.NewBitwiseOrAssign(left, p.parseExpr(Lowest))
}

func (p *Parser) parseBwXorAssign(left ast.Node) ast.Node {
	p.next()
	return ast.NewBitwiseXorAssign(left, p.parseExpr(Lowest))
}

func (p *Parser) parseLShiftAssign(left ast.Node) ast.Node {
	p.next()
	return ast.NewBitwiseShiftLeftAssign(left, p.parseExpr(Lowest))
}

func (p *Parser) parseRShiftAssign(left ast.Node) ast.Node {
	p.next()
	return ast.NewBitwiseShiftRightAssign(left, p.parseExpr(Lowest))
}

func (p *Parser) parseCall(fn ast.Node) ast.Node {
	return ast.NewCall(fn, p.parseNodeList(item.RParen))
}

func (p *Parser) parseIndex(list ast.Node) ast.Node {
	p.next()
	expr := p.parseExpr(Lowest)
	if !p.expectPeek(item.RBracket) {
		return nil
	}
	return ast.NewIndex(list, expr)
}

func (p *Parser) parseDot(left ast.Node) ast.Node {
	prec := p.precedence()
	p.next()
	return ast.NewDot(left, p.parseExpr(prec))
}

func (p *Parser) parsePair() [2]ast.Node {
	l := p.parseExpr(Lowest)
	if !p.expectPeek(item.Colon) {
		return [2]ast.Node{}
	}
	p.next()
	r := p.parseExpr(Lowest)

	return [2]ast.Node{l, r}
}

func (p *Parser) parseNodePairs(end item.Type) [][2]ast.Node {
	var pairs [][2]ast.Node

	p.next()
	if p.cur.Is(end) {
		return pairs
	}

	pairs = append(pairs, p.parsePair())
	for p.peek.Is(item.Comma) {
		p.next()
		p.next()
		pairs = append(pairs, p.parsePair())
	}

	if !p.expectPeek(end) {
		return nil
	}

	return pairs
}

func (p *Parser) parseNodeList(end item.Type) []ast.Node {
	return p.parseNodeSequence(item.Comma, end)
}

// Returns a slice of expressions separated by 'separator'.
func (p *Parser) parseNodeSequence(sep, end item.Type) []ast.Node {
	var seq []ast.Node

	p.next()
	if p.cur.Is(end) {
		return seq
	}

	seq = append(seq, p.parseExpr(Lowest))

	for p.peek.Is(sep) {
		p.next()
		p.next()
		seq = append(seq, p.parseExpr(Lowest))
	}

	if !p.expectPeek(end) {
		return nil
	}
	return seq
}

// Returns true if the peek is of the provided type 't', otherwhise returns
// false and appends an error to p.errs.
func (p *Parser) expectPeek(t item.Type) bool {
	if p.peek.Is(t) {
		p.next()
		return true
	}
	p.peekError(t)
	return false
}

// Emits an error if the peek item is not of tipe t.
func (p *Parser) peekError(t item.Type) {
	p.errs = append(
		p.errs,
		fmt.Sprintf("expected next item to be %v, got %v instead", t, p.peek.Typ),
	)
}

// Returns the precedence value of the type of the peek item.
func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peek.Typ]; ok {
		return prec
	}
	return Lowest
}

// Returns the precedence value of the type of the current item.
func (p *Parser) precedence() int {
	if prec, ok := precedences[p.cur.Typ]; ok {
		return prec
	}
	return Lowest
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
