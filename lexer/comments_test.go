package lexer

import (
	"testing"

	"github.com/esweby/primordial_lang/token"
)

func TestLineComments(t *testing.T) {
	input := `// a whole line
x // after code
// two
// in a row
/ y // at the end`

	expected := []struct {
		typ     token.TokenType
		literal string
		line    int
	}{
		{token.IDENT, "x", 2},
		{token.FORWARD_SLASH, "/", 5},
		{token.IDENT, "y", 5},
		{token.EOF, "", 5},
	}

	l := New(input)
	for i, tt := range expected {
		tok := l.NextToken()
		if tok.Type != tt.typ || tok.Literal != tt.literal || tok.Line != tt.line {
			t.Errorf("token %d: expected %v %q on line %d, got %v %q on line %d",
				i, tt.typ, tt.literal, tt.line, tok.Type, tok.Literal, tok.Line)
		}
	}
}

func TestCommentMarkerInsideAString(t *testing.T) {
	tok := New(`"not // a comment"`).NextToken()
	if tok.Type != token.STRING_LITERAL || tok.Literal != "not // a comment" {
		t.Errorf("expected the whole string, got %v %q", tok.Type, tok.Literal)
	}
}
