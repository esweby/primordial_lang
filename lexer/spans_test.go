package lexer

import (
	"testing"

	"github.com/esweby/primordial_lang/token"
)

func TestTokenSpans(t *testing.T) {
	input := "one := 42\n\"hi\""
	lexer := New(input)

	tests := []struct {
		type_       token.TokenType
		literal     string
		startOffset int
		endOffset   int
		startLine   int
		startColumn int
		endLine     int
		endColumn   int
	}{
		{token.IDENT, "one", 0, 3, 1, 1, 1, 4},
		{token.DECLARE, ":=", 4, 6, 1, 5, 1, 7},
		{token.INT_LITERAL, "42", 7, 9, 1, 8, 1, 10},
		{token.STRING_LITERAL, "hi", 10, 14, 2, 1, 2, 5},
		{token.EOF, "", 14, 14, 2, 5, 2, 5},
	}

	for i, expected := range tests {
		actual := lexer.NextToken()
		if actual.Type != expected.type_ || actual.Literal != expected.literal {
			t.Fatalf("token %d: expected %v %q, got %v %q", i, expected.type_, expected.literal, actual.Type, actual.Literal)
		}
		if actual.Span.Start.Offset != expected.startOffset || actual.Span.End.Offset != expected.endOffset ||
			actual.Span.Start.Line != expected.startLine || actual.Span.Start.Column != expected.startColumn ||
			actual.Span.End.Line != expected.endLine || actual.Span.End.Column != expected.endColumn {
			t.Errorf("token %d: unexpected span: %+v", i, actual.Span)
		}
		if actual.Line != expected.startLine || actual.Column != expected.startColumn {
			t.Errorf("token %d: legacy position does not match span start", i)
		}
	}
}

func TestUnterminatedStringIsIllegal(t *testing.T) {
	lexer := New("\"unfinished")
	tok := lexer.NextToken()

	if tok.Type != token.ILLEGAL {
		t.Fatalf("expected illegal token, got %v", tok.Type)
	}
	if tok.LexError != "unterminated string literal" {
		t.Fatalf("unexpected lexer error: %q", tok.LexError)
	}
	if tok.Span.Start.Offset != 0 || tok.Span.End.Offset != len("\"unfinished") {
		t.Fatalf("unexpected unterminated string span: %+v", tok.Span)
	}
}
