package lexer

import (
	"testing"

	"github.com/NicoNex/tau/item"
)

func TestNextItem(t *testing.T) {
	input := `
five = 5;
ten = 10;

add = fn(x, y) {
  x + y;
};

result = add(five, ten);
! - / * 5 += -= *= /= ** ;
5 < 10 > 5;

if 5 < 10 {
	true;
} else {
	false;
}

10 == 10;
10 != 9;
"foobar";
"foo bar";
[1, 2];

fn mul(x, y) {
	return x * y;
}

10.5 == 10.45;
n = null

`

	tests := []struct {
		expTyp item.Type
		expLit string
	}{
		{item.IDENT, "five"},
		{item.ASSIGN, "="},
		{item.INT, "5"},
		{item.SEMICOLON, ";"},

		{item.IDENT, "ten"},
		{item.ASSIGN, "="},
		{item.INT, "10"},
		{item.SEMICOLON, ";"},

		{item.IDENT, "add"},
		{item.ASSIGN, "="},
		{item.FUNCTION, "fn"},
		{item.LPAREN, "("},
		{item.IDENT, "x"},
		{item.COMMA, ","},
		{item.IDENT, "y"},
		{item.RPAREN, ")"},
		{item.LBRACE, "{"},
		{item.IDENT, "x"},
		{item.PLUS, "+"},
		{item.IDENT, "y"},
		{item.SEMICOLON, ";"},
		{item.RBRACE, "}"},
		{item.SEMICOLON, ";"},

		{item.IDENT, "result"},
		{item.ASSIGN, "="},
		{item.IDENT, "add"},
		{item.LPAREN, "("},
		{item.IDENT, "five"},
		{item.COMMA, ","},
		{item.IDENT, "ten"},
		{item.RPAREN, ")"},
		{item.SEMICOLON, ";"},

		{item.BANG, "!"},
		{item.MINUS, "-"},
		{item.SLASH, "/"},
		{item.ASTERISK, "*"},
		{item.INT, "5"},
		{item.PLUS_ASSIGN, "+="},
		{item.MINUS_ASSIGN, "-="},
		{item.ASTERISK_ASSIGN, "*="},
		{item.SLASH_ASSIGN, "/="},
		{item.POWER, "**"},
		{item.SEMICOLON, ";"},

		{item.INT, "5"},
		{item.LT, "<"},
		{item.INT, "10"},
		{item.GT, ">"},
		{item.INT, "5"},
		{item.SEMICOLON, ";"},

		{item.IF, "if"},
		{item.INT, "5"},
		{item.LT, "<"},
		{item.INT, "10"},
		{item.LBRACE, "{"},
		{item.TRUE, "true"},
		{item.SEMICOLON, ";"},
		{item.RBRACE, "}"},
		{item.ELSE, "else"},
		{item.LBRACE, "{"},
		{item.FALSE, "false"},
		{item.SEMICOLON, ";"},
		{item.RBRACE, "}"},
		{item.SEMICOLON, "\n"},

		{item.INT, "10"},
		{item.EQ, "=="},
		{item.INT, "10"},
		{item.SEMICOLON, ";"},

		{item.INT, "10"},
		{item.NOT_EQ, "!="},
		{item.INT, "9"},
		{item.SEMICOLON, ";"},

		{item.STRING, "foobar"},
		{item.SEMICOLON, ";"},

		{item.STRING, "foo bar"},
		{item.SEMICOLON, ";"},

		{item.LBRACKET, "["},
		{item.INT, "1"},
		{item.COMMA, ","},
		{item.INT, "2"},
		{item.RBRACKET, "]"},
		{item.SEMICOLON, ";"},

		{item.FUNCTION, "fn"},
		{item.IDENT, "mul"},
		{item.LPAREN, "("},
		{item.IDENT, "x"},
		{item.COMMA, ","},
		{item.IDENT, "y"},
		{item.RPAREN, ")"},
		{item.LBRACE, "{"},
		{item.RETURN, "return"},
		{item.IDENT, "x"},
		{item.ASTERISK, "*"},
		{item.IDENT, "y"},
		{item.SEMICOLON, ";"},
		{item.RBRACE, "}"},
		{item.SEMICOLON, "\n"},

		{item.FLOAT, "10.5"},
		{item.EQ, "=="},
		{item.FLOAT, "10.45"},
		{item.SEMICOLON, ";"},

		{item.IDENT, "n"},
		{item.ASSIGN, "="},
		{item.NULL, "null"},
		{item.SEMICOLON, "\n"},

		// {item.LBRACE, "{"},
		// {item.STRING, "foo"},
		// {item.COLON, ":"},
		// {item.STRING, "bar"},
		// {item.RBRACE, "}"},

		{item.EOF, ""},
	}

	items := Lex(input)

	i := 0
	for itm := range items {
		t.Log(itm.Typ, itm.Val)
		if i >= len(tests) {
			break
		}

		tt := tests[i]
		if itm.Typ != tt.expTyp {
			t.Fatalf("tests[%d] - wrong item type: expected=%s, got=%s", i, tt.expTyp, itm.Typ)
		}
		if itm.Val != tt.expLit {
			t.Fatalf("tests[%d] - wrong item literal: expected=%q, got=%q", i, tt.expLit, itm.Val)
		}
		i++
	}
}
