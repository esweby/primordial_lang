package semantic

import (
	"log"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

func TestAnalyzeLoops(t *testing.T) {
	tests := analysisTests{
		{`for {}`, 0},
		{`for (1 < 2) {}`, 0},
		{`for (x := 1; x < 10; x = x + 1) {}`, 0},
		{`for a, b := range map[int32]string{} {}`, 0},
		{`for { break; }`, 0},
		{`lbl: for { break lbl; }`, 0},
		{`x := lbl: for { break lbl; }`, 0},
		{`x := lbl: for { break lbl (1); }`, 0},
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
				log.Printf("test number %d: %s", i, msg)
			}
			t.Fatalf("errors contain %d errors. expected=%d", len(errors), test.expectedErrors)
		}
	}
}
