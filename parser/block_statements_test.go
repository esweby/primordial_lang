package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/lexer"
)

func TestBlockStatementsEndAtTheirClosingBrace(t *testing.T) {
	tests := []struct {
		input      string
		statements int
	}{
		{"if (a) { b; }\n-x;", 2},
		{"if (a) { b; } else { c; }\n-x;", 2},
		{"if (a) { b; }\n(x, y) = f();", 2},
		{"for { break; }\n-x;", 2},
		{"for (a) { b; }\n[]int{1};", 2},
		{"x := if (a) { 1 } else { 2 } + 3;", 1},
	}

	for _, tt := range tests {
		p := New(lexer.New(tt.input))
		program := p.ParseProgram()
		if errs := p.Errors(); len(errs) > 0 {
			t.Fatalf("%q: parser errors: %v", tt.input, errs)
		}
		if len(program.Statements) != tt.statements {
			t.Errorf("%q: expected %d statements, got %d", tt.input, tt.statements, len(program.Statements))
		}
	}
}
