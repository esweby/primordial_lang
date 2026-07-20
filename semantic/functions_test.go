package semantic

import (
	"log"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

func TestAnalyzeFunctionLiteral(t *testing.T) {
	tests := analysisTests{
		{`add := fn() {}`, 0},
		{`add: int32 := fn() {}`, 1},
		{`add: function := fn(x int32, y int32): int32 { return x + y; }`, 0},
		{`add := fn(x int32, y int32): int32 { return x + y; }`, 0},
		{`add := fn(x int32, y int32): int32 { if(x > y) { return y + x;} else { return x; } }`, 0},
	}

	for _, test := range tests {
		l := lexer.New(test.input)
		p := parser.New(l)
		program := p.ParseProgram()

		symbols := NewSymbolTable()
		a := NewSemanticAnalyzer(program, symbols)

		errors := a.Analyze()

		if len(errors) != test.expectedErrors {
			for i, msg := range errors {
				log.Printf("test number %d: %s", i, msg)
			}
			t.Fatalf("errors contain %d errors. expected=%d", len(errors), test.expectedErrors)
		}
	}
}

func TestAnalyzeFunctionStatement(t *testing.T) {
	tests := analysisTests{
		{`fn add() {}`, 0},
		{`fn add(x int32, y int32): int32 { return x + y; }`, 0},
		{`fn add(x int32, y int32): int32 { return x + true; }`, 2},
		{`fn add(x int32, y int32): int32, bool { return x + y, true; }`, 0},
		{`fn add(x int32, y int32): int32, bool { return x + y; }`, 1},
		{`fn add(x int32, y int32) { return; }`, 0},
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
