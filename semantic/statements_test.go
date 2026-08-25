package semantic

import (
	"log"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

func TestAnalyzeDeclareStatement(t *testing.T) {
	tests := analysisTests{
		{`brian := 1;`, 0},
		{`brian := 1; brian := 1`, 1},
		{`brian: int32 := true; `, 1},
		{`brian := if(2 > 1) { 1 + 2; } else { 1 - 0 }`, 0},
		{`brian := if(2) { 1 + 2 } else { 1 - 0 }`, 1},
		{`brian := if(2 > 1) { return 1 + 2; } else { 1 - 0 }`, 1},
		{`brian := if(2 > 1) { 1; } else { true; }`, 1},
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

func TestAssignmentAnalysis(t *testing.T) {
	tests := analysisTests{
		{`mut brian := 1; brian = 2`, 0},
		{
			`struct Person { 
				pub name: string; 
				fn new(): Person { 
					return Person{ name: "jenny" };
				}
			 }; 
			 p := Person.new(); 
			 p.name = "Bob";`,
			0,
		},
		{`x := map[int32]string{}; x[12] = "one";`, 0},
		{`x := map[string]int{}; x["key"] = 42;`, 0},
		{`x := map[bool]string{}; x[true] = "yes";`, 0},
		{`mut i := int32(1); x := map[int32]string{}; x[i] = "one";`, 0},
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
				log.Printf("test %d: %s", i, msg)
			}
			t.Fatalf("test %d errors contain %d errors. expected=%d", i, len(errors), test.expectedErrors)
		}
	}
}
