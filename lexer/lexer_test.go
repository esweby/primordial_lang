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
			t.Fatalf("tests[%d] - tokentype wrong, expected=%q got=%q", i, token.GetTokenName(int(tt.expectedType)), token.GetTokenName(int(tok.Type)))
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
	input := `<mut, pub, int> five := 5;`

	tests := Tests{
		{token.LTAG, "<"},
		{token.MUT, "mut"},
		{token.COMMA, ","},
		{token.PUB, "pub"},
		{token.COMMA, ","},
		{token.IDENT, "int"},
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
		{token.IDENT, "int"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.IDENT, "int"},
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

func TestBooleanTokens(t *testing.T) {
	input := `isThis := true;
isThat := false;`

	tests := Tests{
		{token.IDENT, "isThis"},
		{token.DECLARE, ":="},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "isThat"},
		{token.DECLARE, ":="},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
	}

	checkTokens(t, input, tests)
}

func TestComparisonOperators(t *testing.T) {
	input := `x == y;
x != y;
x > y;
x < y;
x >= y;
x <= y;`

	tests := Tests{
		{token.IDENT, "x"},
		{token.EQUALS, "=="},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "x"},
		{token.NOT_EQUALS, "!="},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "x"},
		{token.RTAG, ">"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "x"},
		{token.LTAG, "<"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "x"},
		{token.GREATER_THAN_OR_EQUALS, ">="},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.IDENT, "x"},
		{token.LESS_THAN_OR_EQUALS, "<="},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
	}

	checkTokens(t, input, tests)
}

func TestBranchingToken(t *testing.T) {
	input := `if (x < y) {
	return true;	
} else {
	return false;
}`

	tests := Tests{
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.LTAG, "<"},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.RBRACE, "}"},
	}

	checkTokens(t, input, tests)
}

func TestStringLiteral(t *testing.T) {
	input := `
	"foobar"; 
	"footbar";
	`

	tests := Tests{
		{token.STRING_LITERAL, "foobar"},
		{token.SEMICOLON, ";"},
		{token.STRING_LITERAL, "footbar"},
		{token.SEMICOLON, ";"},
	}

	checkTokens(t, input, tests)
}

func TestArrayLiteral(t *testing.T) {
	input := `
	x: [3]int32 := [3]int32{1, 2, 3};
	`

	tests := Tests{
		{token.IDENT, "x"},
		{token.COLON, ":"},
		{token.LBRACKET, "["},
		{token.INT_LITERAL, "3"},
		{token.RBRACKET, "]"},
		{token.IDENT, "int32"},
		{token.DECLARE, ":="},
		{token.LBRACKET, "["},
		{token.INT_LITERAL, "3"},
		{token.RBRACKET, "]"},
		{token.IDENT, "int32"},
		{token.LBRACE, "{"},
		{token.INT_LITERAL, "1"},
		{token.COMMA, ","},
		{token.INT_LITERAL, "2"},
		{token.COMMA, ","},
		{token.INT_LITERAL, "3"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
	}

	checkTokens(t, input, tests)
}
