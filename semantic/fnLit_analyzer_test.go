package semantic

import (
	"log"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

func TestFunctionLiteralAnalysis(t *testing.T) {
	tests := Tests{
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

		a := NewSemanticAnalyzer(program)
		errors := a.Analyze()

		if len(errors) != test.numErrors {
			for i, msg := range errors {
				log.Printf("test number %d: %s", i, msg)
			}
			t.Fatalf("errors contain %d errors. expected=%d", len(errors), test.numErrors)
		}
	}
}
