package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
)

func TestParsingArrayLiteral(t *testing.T) {
	tests := []struct{
		input string
		expectedElements []int64
	}{
		{`input: [3]int32 := [3]int32{1, 2, 3};`, []int64{1, 2, 3}},
		{`input: [3]int32 := {1, 2, 3}`, []int64{1, 2, 3}},
		{`input := [3]int32{1, 2, 3}`, []int64{1, 2, 3}},
	}

	for i, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		requireNoParserErrors(t, p)
		requireStatementCount(t, program.Statements, 1)

		declaration, ok := program.Statements[0].(*ast.DeclareStatement)
		if !ok {
			t.Fatalf("test %d: expected declareStatement, got %T", i, program.Statements[0])
		}

		literal, ok := declaration.Value.(*ast.ArrayLiteral)
		if !ok {
			t.Fatalf("test %d: expected arrayLiteral, got %T", i, declaration.Value)
		}

		if len(literal.Elements) != len(tt.expectedElements) {
			t.Fatalf("test %d: expected %d elements, got %d", i, len(tt.expectedElements), len(literal.Elements))
		}

		for vi, v := range literal.Elements {
			testIntegerLiteral(t, v, int64(tt.expectedElements[vi]))
		}
	}
}

func TestParseTupleDeclarationAndAssignment(t *testing.T) {
	input := `(first, _) := getNames(); (first, last) = getNames();`
	p := New(lexer.New(input))
	program := p.ParseProgram()
	requireNoParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}

	declaration, ok := program.Statements[0].(*ast.TupleDeclareStatement)
	if !ok {
		t.Fatalf("expected tuple declaration, got %T", program.Statements[0])
	}
	if declaration.Names[0].Value != "first" || declaration.Names[1].Value != "_" {
		t.Fatalf("unexpected tuple declaration names: %s", declaration.String())
	}

	assignment, ok := program.Statements[1].(*ast.TupleAssignStatement)
	if !ok {
		t.Fatalf("expected tuple assignment, got %T", program.Statements[1])
	}
	if assignment.Names[0].Value != "first" || assignment.Names[1].Value != "last" {
		t.Fatalf("unexpected tuple assignment names: %s", assignment.String())
	}
}

func TestRejectTypedTupleDeclaration(t *testing.T) {
	p := New(lexer.New(`(first: int32, last) := getNames();`))
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Fatal("expected typed tuple declaration to produce a parser error")
	}
}
