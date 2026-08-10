package parser

import (
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/token"
)

func TestStructuredFunctionParameterDiagnostic(t *testing.T) {
	parser := New(lexer.New("fn add(x int32 y int32) {}"))
	parser.ParseProgram()

	diagnostics := parser.Diagnostics()
	if len(diagnostics) == 0 {
		t.Fatal("expected parser diagnostic")
	}

	diagnostic := diagnostics[0]
	if diagnostic.Code != "P1203" {
		t.Fatalf("expected P1203, got %s (%s)", diagnostic.Code, diagnostic.Message)
	}
	if diagnostic.Found.Type != token.IDENT || diagnostic.Found.Literal != "y" {
		t.Fatalf("expected offending token y, got %+v", diagnostic.Found)
	}
	if diagnostic.Span.Start.Line != 1 || diagnostic.Span.Start.Column != 16 {
		t.Fatalf("expected diagnostic at 1:16, got %+v", diagnostic.Span.Start)
	}
	if len(diagnostic.Expected) != 2 || diagnostic.Expected[0] != token.COMMA || diagnostic.Expected[1] != token.RPAREN {
		t.Fatalf("unexpected expected tokens: %v", diagnostic.Expected)
	}
	if !strings.HasPrefix(diagnostic.Error(), "1:16 [P1203]") {
		t.Fatalf("unexpected formatted diagnostic: %s", diagnostic.Error())
	}
}

func TestPreviouslySilentParserFailuresProduceDiagnostics(t *testing.T) {
	tests := []struct {
		name  string
		input string
		code  string
	}{
		{name: "declaration name", input: "mut := 1;", code: "P1101"},
		{name: "declaration operator", input: "value: int32 = 1;", code: "P1102"},
		{name: "function name", input: "fn (x int32) {}", code: "P1201"},
		{name: "function opening parenthesis", input: "fn value {}", code: "P1202"},
		{name: "struct name", input: "struct {}", code: "P1301"},
		{name: "struct field type", input: "struct Value { field: ; }", code: "P1401"},
		{name: "impl opening brace", input: "struct Value { impl fn broken() {} }", code: "P1312"},
		{name: "collection element type", input: "value := []{};", code: "P1504"},
		{name: "return terminator", input: "fn value(): int32 { return 1 }", code: "P1702"},
		{name: "closing block", input: "fn value() { 1;", code: "P1008"},
		{name: "unterminated string", input: "value := \"unfinished", code: "L1001"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			parser := New(lexer.New(test.input))
			parser.ParseProgram()
			diagnostics := parser.Diagnostics()
			if len(diagnostics) == 0 {
				t.Fatal("expected parser diagnostic")
			}
			found := false
			for _, diagnostic := range diagnostics {
				if diagnostic.Code == test.code {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected diagnostic %s, got %+v", test.code, diagnostics)
			}
		})
	}
}

func TestLexicalDiagnosticRetainsItsPhase(t *testing.T) {
	parser := New(lexer.New("value := \"unfinished"))
	parser.ParseProgram()
	diagnostics := parser.Diagnostics()
	if len(diagnostics) == 0 {
		t.Fatal("expected lexical diagnostic")
	}
	if diagnostics[0].Phase != PhaseLexing || diagnostics[0].Message != "unterminated string literal" {
		t.Fatalf("unexpected lexical diagnostic: %+v", diagnostics[0])
	}
}
