package semantic

import (
	"log"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

func TestCallExpressionAnalysis(t *testing.T) {
	tests := Tests{
		 {`fn add(x int32): int32 { return 10 + 1; } add(5);`, 0},
		 {`fn add(x int32): int32 { return x + 1; } add(true);`, 1},
	}

	for i, test := range tests {
		l := lexer.New(test.input)
		p := parser.New(l)
		program := p.ParseProgram()

		a := NewSemanticAnalyzer(program)
		errors := a.Analyze()

		if len(errors) != test.numErrors {
			for _, msg := range errors {
				log.Printf("%s", msg)
			}
			t.Fatalf("test %d errors contain %d errors. expected=%d", i, len(errors), test.numErrors)
		}
	}
}
