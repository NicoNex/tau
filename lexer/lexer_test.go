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
		{item.Assign, "="},
		{item.Int, "5"},
		{item.Semicolon, ";"},

		{item.IDENT, "ten"},
		{item.Assign, "="},
		{item.Int, "10"},
		{item.Semicolon, ";"},

		{item.IDENT, "add"},
		{item.Assign, "="},
		{item.FUNCTION, "fn"},
		{item.LParen, "("},
		{item.IDENT, "x"},
		{item.Comma, ","},
		{item.IDENT, "y"},
		{item.RParen, ")"},
		{item.LBrace, "{"},
		{item.IDENT, "x"},
		{item.Plus, "+"},
		{item.IDENT, "y"},
		{item.Semicolon, ";"},
		{item.RBrace, "}"},
		{item.Semicolon, ";"},

		{item.IDENT, "result"},
		{item.Assign, "="},
		{item.IDENT, "add"},
		{item.LParen, "("},
		{item.IDENT, "five"},
		{item.Comma, ","},
		{item.IDENT, "ten"},
		{item.RParen, ")"},
		{item.Semicolon, ";"},

		{item.Bang, "!"},
		{item.Minus, "-"},
		{item.Slash, "/"},
		{item.Asterisk, "*"},
		{item.Int, "5"},
		{item.PlusAssign, "+="},
		{item.MinusAssign, "-="},
		{item.AsteriskAssign, "*="},
		{item.SlashAssign, "/="},
		{item.Power, "**"},
		{item.Semicolon, ";"},

		{item.Int, "5"},
		{item.LT, "<"},
		{item.Int, "10"},
		{item.GT, ">"},
		{item.Int, "5"},
		{item.Semicolon, ";"},

		{item.IF, "if"},
		{item.Int, "5"},
		{item.LT, "<"},
		{item.Int, "10"},
		{item.LBrace, "{"},
		{item.TRUE, "true"},
		{item.Semicolon, ";"},
		{item.RBrace, "}"},
		{item.ELSE, "else"},
		{item.LBrace, "{"},
		{item.FALSE, "false"},
		{item.Semicolon, ";"},
		{item.RBrace, "}"},
		{item.Semicolon, "\n"},

		{item.Int, "10"},
		{item.Equals, "=="},
		{item.Int, "10"},
		{item.Semicolon, ";"},

		{item.Int, "10"},
		{item.NotEquals, "!="},
		{item.Int, "9"},
		{item.Semicolon, ";"},

		{item.String, "foobar"},
		{item.Semicolon, ";"},

		{item.String, "foo bar"},
		{item.Semicolon, ";"},

		{item.LBracket, "["},
		{item.Int, "1"},
		{item.Comma, ","},
		{item.Int, "2"},
		{item.RBracket, "]"},
		{item.Semicolon, ";"},

		{item.FUNCTION, "fn"},
		{item.IDENT, "mul"},
		{item.LParen, "("},
		{item.IDENT, "x"},
		{item.Comma, ","},
		{item.IDENT, "y"},
		{item.RParen, ")"},
		{item.LBrace, "{"},
		{item.RETURN, "return"},
		{item.IDENT, "x"},
		{item.Asterisk, "*"},
		{item.IDENT, "y"},
		{item.Semicolon, ";"},
		{item.RBrace, "}"},
		{item.Semicolon, "\n"},

		{item.Float, "10.5"},
		{item.Equals, "=="},
		{item.Float, "10.45"},
		{item.Semicolon, ";"},

		{item.IDENT, "n"},
		{item.Assign, "="},
		{item.NULL, "null"},
		{item.Semicolon, "\n"},

		// {item.LBrace, "{"},
		// {item.String, "foo"},
		// {item.Colon, ":"},
		// {item.String, "bar"},
		// {item.RBrace, "}"},

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
