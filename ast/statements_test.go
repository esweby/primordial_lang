package ast

import (
	"testing"

	"github.com/esweby/primordial_lang/token"
)

func TestDeclareStatementString(t *testing.T) {
	program := &Program{
		Statements: []Statement{
			&DeclareStatement{
				Token: token.Token{Type: token.DECLARE, Literal: ":="},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "anotherVar"},
					Value: "anotherVar",
				},
			},
		},
	}

	if program.String() != "myVar := anotherVar;" {
		t.Errorf("program.String() wrong. got=%q", program.String())
	}
}
