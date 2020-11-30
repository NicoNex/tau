package item

type Type int

const (
	EOF Type = iota
	ERROR

	IDENT
	INT
	FLOAT
	STRING

	ASSIGN
	PLUS
	MINUS
	SLASH
	ASTERISK
	POWER
	EQ
	NOT_EQ
	BANG
	LT
	GT
	LT_EQ
	GT_EQ
	AND
	OR

	COMMA
	SEMICOLON
	NEW_LINE

	LPAREN
	RPAREN

	LBRACE
	RBRACE

	LBRACKET
	RBRACKET

	FUNCTION
	IF
	ELSE
	TRUE
	FALSE
)

var typemap = map[Type]string{
	EOF:   "EOF",
	ERROR: "ERROR",

	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	STRING: "STRING",

	ASSIGN:   "=",
	PLUS:     "+",
	MINUS:    "*",
	SLASH:    "/",
	ASTERISK: "*",
	POWER:    "**",
	EQ:       "==",
	NOT_EQ:   "!=",
	BANG:     "!",
	LT:       "<",
	GT:       ">",
	LT_EQ:    "<=",
	GT_EQ:    ">=",
	AND:      "&&",
	OR:       "||",

	COMMA:     ",",
	SEMICOLON: ";",

	LPAREN:  "(",
	RPAREN:  ")",

	LBRACE: "{",
	RBRACE: "}",

	LBRACKET: "[",
	RBRACKET: "]",

	FUNCTION: "FUNCTION",
	IF:       "IF",
	ELSE:     "ELSE",
	TRUE:     "TRUE",
	FALSE:    "FALSE",
}

var keywords = map[string]Type{
	"fn":     FUNCTION,
	"if":     IF,
	"else":   ELSE,
	"true":   TRUE,
	"false":  FALSE,
}

func (t Type) String() string {
	return typemap[t]
}

func Lookup(ident string) Type {
	if t, ok := keywords[ident]; ok {
		return t
	}
	return IDENT
}
