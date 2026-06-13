package semantic

import (
	"log"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

type fnTestToken struct {
	input     string
	numErrors int
}

type fnTests []fnTestToken

func TestFunctionAnalysis(t *testing.T) {
	tests := fnTests{
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

		a := New(program)
		errors := a.Analyze()

		if len(errors) != test.numErrors {
			for _, msg := range errors {
				log.Printf("%s", msg)
			}
			t.Fatalf("errors contain %d errors. expected=%d", len(errors), test.numErrors)
		}
	}
}
