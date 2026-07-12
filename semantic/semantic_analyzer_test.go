package semantic

import (
	"log"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

type TestToken struct {
	input     string
	numErrors int
}

type Tests []TestToken

func TestInfixExpression(t *testing.T) {
	tests := Tests{
		{`2 + 2 + 2 + 2`, 0},
		{`true + 2`, 1},
		{`2 * 2`, 0},
		{`true * 2`, 1},
		{`1 + true`, 1},
		{`1 + fn() {}`, 1},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := parser.New(l)
		program := p.ParseProgram()

		symbols := NewSymbolTable()
		a := NewSemanticAnalyzer(program, symbols)

		errors := a.Analyze()

		if len(errors) != test.numErrors {
			t.Fatalf("errors contain %d errors. expected=%d", len(errors), test.numErrors)
		}
	}
}

func TestPrefixExpression(t *testing.T) {
	tests := Tests{
		{`!5`, 1},
		{`!true`, 0},
		{`-true`, 1},
		{`-10`, 0},
	}

	for _, test := range tests {
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
			t.Fatalf("errors contain %d errors. expected=%d", len(errors), test.numErrors)
		}
	}
}
