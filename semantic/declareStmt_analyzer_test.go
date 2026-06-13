package semantic

import (
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

type dsTestToken struct {
	input     string
	numErrors int
}

type dsTests []dsTestToken

func TestDeclareAnalysis(t *testing.T) {
	tests := dsTests{
		{`brian := 1;`, 0},
		{`brian := 1; brian := 1`, 1},
		{`brian: int32 := true; `, 1},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := parser.New(l)
		program := p.ParseProgram()

		a := New(program)
		errors := a.Analyze()

		if len(errors) != test.numErrors {
			t.Fatalf("errors contain %d errors. expected=%d", len(errors), test.numErrors)
		}
	}
}
