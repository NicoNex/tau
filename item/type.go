package item

type Type int

const (
	EOF Type = iota
	ERROR
	NULL

	IDENT
	INT
	FLOAT
	STRING

	ASSIGN
	PLUS
	MINUS
	SLASH
	ASTERISK
	MODULUS
	PLUS_ASSIGN
	MINUS_ASSIGN
	ASTERISK_ASSIGN
	SLASH_ASSIGN
	MODULUS_ASSIGN
	BWAND_ASSIGN
	BWOR_ASSIGN
	BWXOR_ASSIGN
	LSHIFT_ASSIGN
	RSHIFT_ASSIGN
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
	IN
	BWAND
	BWOR
	BWXOR
	LSHIFT
	RSHIFT
	PLUSPLUS
	MINUSMINUS

	DOT
	COMMA
	COLON
	SEMICOLON
	NEW_LINE

	LPAREN
	RPAREN

	LBRACE
	RBRACE

	LBRACKET
	RBRACKET

	FUNCTION
	FOR
	IF
	ELSE
	TRUE
	FALSE
	RETURN
)

var typemap = map[Type]string{
	EOF:   "eof",
	ERROR: "error",
	NULL:  "null",

	IDENT:  "IDENT",
	INT:    "int",
	FLOAT:  "float",
	STRING: "string",

	ASSIGN:          "=",
	PLUS:            "+",
	MINUS:           "*",
	SLASH:           "/",
	ASTERISK:        "*",
	MODULUS:         "%",
	POWER:           "**",
	EQ:              "==",
	NOT_EQ:          "!=",
	BANG:            "!",
	LT:              "<",
	GT:              ">",
	LT_EQ:           "<=",
	GT_EQ:           ">=",
	AND:             "&&",
	OR:              "||",
	IN:              "in",
	PLUS_ASSIGN:     "+=",
	MINUS_ASSIGN:    "-=",
	ASTERISK_ASSIGN: "*=",
	SLASH_ASSIGN:    "/=",
	PLUSPLUS:        "++",
	MINUSMINUS:      "--",
	BWAND:           "&",
	BWOR:            "|",
	BWXOR:           "^",
	LSHIFT:          "<<",
	RSHIFT:          ">>",

	DOT:       ".",
	COMMA:     ",",
	COLON:     ":",
	SEMICOLON: ";",
	NEW_LINE:  "new line",

	LPAREN: "(",
	RPAREN: ")",

	LBRACE: "{",
	RBRACE: "}",

	LBRACKET: "[",
	RBRACKET: "]",

	FUNCTION: "function",
	FOR:      "for",
	IF:       "if",
	ELSE:     "else",
	TRUE:     "true",
	FALSE:    "false",
	RETURN:   "return",
}

var keywords = map[string]Type{
	"in":     IN,
	"fn":     FUNCTION,
	"for":    FOR,
	"if":     IF,
	"else":   ELSE,
	"true":   TRUE,
	"false":  FALSE,
	"return": RETURN,
	"null":   NULL,
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
