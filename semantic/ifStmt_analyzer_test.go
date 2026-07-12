package semantic

import (
	"log"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

func TestIfStatementAnalysis(t *testing.T) {
	tests := Tests{
		{`if(1 > 2) { 1; 2; 3; };`, 0},
		{`if(x > y) { 1; 2; 3; };`, 3},
		{`x := 1; y := 2; if(1 > 2) { 1; 2; 3; };`, 0},
	}

	for i, test := range tests {
		l := lexer.New(test.input)
		p := parser.New(l)
		program := p.ParseProgram()

		symbols := NewSymbolTable()
		a := NewSemanticAnalyzer(program, symbols)

		errors := a.Analyze()

		if len(errors) != test.numErrors {
			for _, msg := range errors {
				log.Printf("%s", msg)
			}
			t.Fatalf("test %d errors contain %d errors. expected=%d", i, len(errors), test.numErrors)
		}
	}
}
