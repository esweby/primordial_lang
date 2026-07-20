package semantic

import (
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

func analyzeTupleInput(input string) []error {
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	return NewSemanticAnalyzer(program, NewSymbolTable()).Analyze()
}

func TestAnalyzeTupleDeclaration(t *testing.T) {
	errors := analyzeTupleInput(`
		fn values(): int32, bool { return 10, true; }
		(number, _) := values();
		number;
	`)
	if len(errors) != 0 {
		t.Fatalf("expected tuple declaration to analyze, got %v", errors)
	}
}

func TestAnalyzeTupleAssignment(t *testing.T) {
	errors := analyzeTupleInput(`
		fn values(): int32, int32 { return 10, 20; }
		mut first := 0;
		mut second := 0;
		(first, second) = values();
	`)
	if len(errors) != 0 {
		t.Fatalf("expected tuple assignment to analyze, got %v", errors)
	}
}

func TestRejectImmutableTupleAssignment(t *testing.T) {
	errors := analyzeTupleInput(`
		fn values(): int32, int32 { return 10, 20; }
		first := 0;
		mut second := 0;
		(first, second) = values();
	`)
	if len(errors) == 0 {
		t.Fatal("expected assignment to an immutable tuple target to fail")
	}
}
