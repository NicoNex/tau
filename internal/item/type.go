package item

type Type int

const (
	EOF Type = iota
	Error
	Null

	Ident
	Int
	Float
	String
	RawString

	Assign
	Plus
	Minus
	Slash
	Asterisk
	Modulus
	PlusAssign
	MinusAssign
	AsteriskAssign
	SlashAssign
	ModulusAssign
	BwAndAssign
	BwOrAssign
	BwXorAssign
	LShiftAssign
	RShiftAssign
	Power
	Equals
	NotEquals
	Bang
	LT
	GT
	LTEQ
	GTEQ
	And
	Or
	In
	BwAnd
	BwNot
	BwOr
	BwXor
	LShift
	RShift
	PlusPlus
	MinusMinus

	Dot
	Comma
	Colon
	Semicolon
	NewLine

	LParen
	RParen

	LBrace
	RBrace

	LBracket
	RBracket

	Function
	For
	Continue
	Break
	If
	Else
	True
	False
	Return
	Import
	Tau
)

var typemap = map[Type]string{
	EOF:   "eof",
	Error: "error",
	Null:  "null",

	Ident:     "IDENT",
	Int:       "int",
	Float:     "float",
	String:    "string",
	RawString: "rawstring",

	Assign:         "=",
	Plus:           "+",
	Minus:          "*",
	Slash:          "/",
	Asterisk:       "*",
	Modulus:        "%",
	Power:          "**",
	Equals:         "==",
	NotEquals:      "!=",
	Bang:           "!",
	LT:             "<",
	GT:             ">",
	LTEQ:           "<=",
	GTEQ:           ">=",
	And:            "&&",
	Or:             "||",
	In:             "in",
	PlusAssign:     "+=",
	MinusAssign:    "-=",
	AsteriskAssign: "*=",
	SlashAssign:    "/=",
	ModulusAssign:  "%=",
	BwAndAssign:    "&=",
	BwOrAssign:     "|=",
	BwXorAssign:    "^=",
	LShiftAssign:   "<<=",
	RShiftAssign:   ">>=",
	PlusPlus:       "++",
	MinusMinus:     "--",
	BwAnd:          "&",
	BwNot:          "~",
	BwOr:           "|",
	BwXor:          "^",
	LShift:         "<<",
	RShift:         ">>",

	Dot:       ".",
	Comma:     ",",
	Colon:     ":",
	Semicolon: ";",
	NewLine:   "new line",

	LParen: "(",
	RParen: ")",

	LBrace: "{",
	RBrace: "}",

	LBracket: "[",
	RBracket: "]",

	Function: "function",
	For:      "for",
	Continue: "continue",
	Break:    "break",
	If:       "if",
	Else:     "else",
	True:     "true",
	False:    "false",
	Return:   "return",
	Import:   "import",
}

var keywords = map[string]Type{
	"in":       In,
	"fn":       Function,
	"for":      For,
	"continue": Continue,
	"break":    Break,
	"if":       If,
	"else":     Else,
	"true":     True,
	"false":    False,
	"return":   Return,
	"null":     Null,
	"import":   Import,
	"tau":      Tau,
}

func (t Type) String() string {
	return typemap[t]
}

func Lookup(ident string) Type {
	if t, ok := keywords[ident]; ok {
		return t
	}
	return Ident
}
