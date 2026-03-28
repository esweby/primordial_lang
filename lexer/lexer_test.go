package lexer

import (
	"testing"

	"github.com/esweby/primordial_lang/token"
)

type TestToken struct {
	expectedType    token.TokenType
	expectedLiteral string
}

type Tests []TestToken

func checkTokens(t *testing.T, input string, expected Tests) {
	l := New(input)

	for i, tt := range expected {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q got=%q", i, tt.expectedType, tok.ToString())
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - token literal wrong, expected=%q got=%q", i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken(t *testing.T) {
	input := `=+(){},;`

	tests := Tests{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.LPAREN, "("},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RBRACE, "}"},
		{token.COMMA, ","},
		{token.SEMICOLON, ";"},
	}

	checkTokens(t, input, tests)
}

func TestDeclareToken(t *testing.T) {
	input := `five := 5`

	tests := Tests{
		{token.IDENT, "five"},
		{token.DECLARE, ":="},
		{token.INT_LITERAL, "5"},
		{token.EOF, ""},
	}

	checkTokens(t, input, tests)
}

func TestDeclareWithOptions(t *testing.T) {
	input := `<mut, exp, int> five := 5;`

	tests := Tests{
		{token.LTAG, "<"},
		{token.MUT, "mut"},
		{token.COMMA, ","},
		{token.EXP, "exp"},
		{token.COMMA, ","},
		{token.INT_LITERAL, "int"},
		{token.RTAG, ">"},
		{token.IDENT, "five"},
		{token.DECLARE, ":="},
		{token.INT_LITERAL, "5"},
		{token.SEMICOLON, ";"},
	}

	checkTokens(t, input, tests)
}

func TestAssignToken(t *testing.T) {
	input := `five = 5`

	tests := Tests{
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT_LITERAL, "5"},
		{token.EOF, ""},
	}

	checkTokens(t, input, tests)
}

func TestDeclareAndAssign(t *testing.T) {
	input := `ten := 5;
ten = ten + 5;`

	tests := Tests{
		{token.IDENT, "ten"},
		{token.DECLARE, ":="},
		{token.INT_LITERAL, "5"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.IDENT, "ten"},
		{token.PLUS, "+"},
		{token.INT_LITERAL, "5"},
		{token.SEMICOLON, ";"},
		{token.EOF, ""},
	}

	checkTokens(t, input, tests)
}

func TestFunctionDeclare(t *testing.T) {
	input := `fn add(x int, y int) {
	return x + y;	
}`

	tests := Tests{
		{token.FN, "fn"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.INT_LITERAL, "int"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.INT_LITERAL, "int"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.IDENT, "x"}, 
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
	}

	checkTokens(t, input, tests)
}