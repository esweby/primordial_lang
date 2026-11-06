package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
)

func TestBreakWithoutASemicolon(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`for { break }`, "break"},
		{`a: for { break a }`, "break a"},
		{`for { break (1) }`, "break (1)"},
		{`for { break; }`, "break"},
		{`a: for { break a (1); }`, "break a (1)"},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))
		program := p.ParseProgram()
		if errs := p.Errors(); len(errs) > 0 {
			t.Fatalf("%s: parser errors: %v", tt.input, errs)
		}
		loop := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.ForLoop)
		if len(loop.Body.Statements) != 1 {
			t.Fatalf("%s: expected 1 statement, got %d", tt.input, len(loop.Body.Statements))
		}
		if actual := loop.Body.Statements[0].String(); actual != tt.expected {
			t.Errorf("%s: expected %q, got %q", tt.input, tt.expected, actual)
		}
	}
}
