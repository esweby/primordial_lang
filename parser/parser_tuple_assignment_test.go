package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
)

func TestTupleDeclarationAndAssignment(t *testing.T) {
	input := `(first, _) := getNames(); (first, last) = getNames();`
	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

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

func TestTupleDeclarationRejectsTypes(t *testing.T) {
	p := New(lexer.New(`(first: int32, last) := getNames();`))
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Fatal("expected typed tuple declaration to produce a parser error")
	}
}
