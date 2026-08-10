package semantic

import (
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
)

func analyzeTupleInput(input string) []error {
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	return NewSemanticAnalyzer(program, NewSymbolTable()).Analyze()
}

func analyzeCollectionInput(
	t *testing.T,
	input string,
) (*ast.Program, *SemanticAnalyzer, []error) {
	t.Helper()

	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}

	analyzer := NewSemanticAnalyzer(program, NewSymbolTable())
	return program, analyzer, analyzer.Analyze()
}

func TestAnalyzePartiallyInitializedArray(t *testing.T) {
	program, analyzer, errors := analyzeCollectionInput(
		t,
		`values := [5]int32{1, 2};`,
	)
	if len(errors) != 0 {
		t.Fatalf("expected array to analyze, got %v", errors)
	}

	symbol, ok := analyzer.Symbols().Get("values")
	if !ok {
		t.Fatal("expected values to be registered")
	}
	if symbol.Type().Name() != "[5]int32" {
		t.Fatalf("expected inferred type [5]int32, got %s", symbol.Type().Name())
	}

	declaration, ok := program.Statements[0].(*ast.DeclareStatement)
	if !ok {
		t.Fatalf("expected declaration, got %T", program.Statements[0])
	}
	literal, ok := declaration.Value.(*ast.ArrayLiteral)
	if !ok {
		t.Fatalf("expected array literal, got %T", declaration.Value)
	}
	if len(literal.Elements) != 2 {
		t.Fatalf("semantic analysis should preserve 2 explicit elements, got %d", len(literal.Elements))
	}
}

func TestAnalyzeEmptyArrayLiteral(t *testing.T) {
	_, analyzer, errors := analyzeCollectionInput(
		t,
		`values := [3]int32{};`,
	)
	if len(errors) != 0 {
		t.Fatalf("expected empty array literal to analyze, got %v", errors)
	}

	symbol, ok := analyzer.Symbols().Get("values")
	if !ok {
		t.Fatal("expected values to be registered")
	}
	if symbol.Type().Name() != "[3]int32" {
		t.Fatalf("expected inferred type [3]int32, got %s", symbol.Type().Name())
	}
}

func TestRejectTooManyArrayElements(t *testing.T) {
	_, _, errors := analyzeCollectionInput(
		t,
		`values := [2]int32{1, 2, 3};`,
	)
	if len(errors) != 1 {
		t.Fatalf("expected one semantic error, got %v", errors)
	}
	if !strings.Contains(errors[0].Error(), "array length is 2") {
		t.Fatalf("expected array length error, got %v", errors[0])
	}
}

func TestRejectInvalidArrayElementType(t *testing.T) {
	_, _, errors := analyzeCollectionInput(
		t,
		`values := [2]int32{1, true};`,
	)
	if len(errors) != 1 {
		t.Fatalf("expected one semantic error, got %v", errors)
	}
	if !strings.Contains(errors[0].Error(), "array element 1") {
		t.Fatalf("expected array element error, got %v", errors[0])
	}
}

func TestRejectMismatchedDeclaredArrayType(t *testing.T) {
	_, _, errors := analyzeCollectionInput(
		t,
		`values: [2]int32 := [3]int32{1};`,
	)
	if len(errors) != 1 {
		t.Fatalf("expected one semantic error, got %v", errors)
	}
	if !strings.Contains(errors[0].Error(), "declaration type mismatch") {
		t.Fatalf("expected declaration mismatch, got %v", errors[0])
	}
}

func TestAnalyzeSliceLiteral(t *testing.T) {
	_, analyzer, errors := analyzeCollectionInput(
		t,
		`values := []int32{1, 2, 3};`,
	)
	if len(errors) != 0 {
		t.Fatalf("expected slice to analyze, got %v", errors)
	}

	symbol, ok := analyzer.Symbols().Get("values")
	if !ok {
		t.Fatal("expected values to be registered")
	}
	if symbol.Type().Name() != "[]int32" {
		t.Fatalf("expected inferred type []int32, got %s", symbol.Type().Name())
	}
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
		mut first: int32 := 0;
		mut second: int32 := 0;
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
