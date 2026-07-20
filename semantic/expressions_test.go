package semantic

import (
	"log"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

type analysisTest struct {
	input          string
	expectedErrors int
}

type analysisTests []analysisTest

func TestAnalyzeInfixExpression(t *testing.T) {
	tests := analysisTests{
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

		if len(errors) != test.expectedErrors {
			t.Fatalf("errors contain %d errors. expected=%d", len(errors), test.expectedErrors)
		}
	}
}

func TestAnalyzePrefixExpression(t *testing.T) {
	tests := analysisTests{
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

		if len(errors) != test.expectedErrors {
			for _, msg := range errors {
				log.Printf("%s", msg)
			}
			t.Fatalf("errors contain %d errors. expected=%d", len(errors), test.expectedErrors)
		}
	}
}

func TestAnalyzeCallExpression(t *testing.T) {
	tests := analysisTests{
		{`fn add(x int32): int32 { return 10 + 1; } add(5);`, 0},
		{`fn add(x int32): int32 { return x + 1; } add(true);`, 1},
	}

	for i, test := range tests {
		l := lexer.New(test.input)
		p := parser.New(l)
		program := p.ParseProgram()
		symbols := NewSymbolTable()
		a := NewSemanticAnalyzer(program, symbols)
		errors := a.Analyze()

		if len(errors) != test.expectedErrors {
			for _, msg := range errors {
				log.Printf("%s", msg)
			}
			t.Fatalf("test %d errors contain %d errors. expected=%d", i, len(errors), test.expectedErrors)
		}
	}
}

func TestAnalyzeIfExpression(t *testing.T) {
	tests := analysisTests{
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

		if len(errors) != test.expectedErrors {
			for _, msg := range errors {
				log.Printf("%s", msg)
			}
			t.Fatalf("test %d errors contain %d errors. expected=%d", i, len(errors), test.expectedErrors)
		}
	}
}
