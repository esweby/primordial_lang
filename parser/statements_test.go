package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
)

func TestParseDeclareStatements(t *testing.T) {
	input := `x := 5;
y := 10;
mut cats := 12;
pub dogs := 5;
pub mut rats := 10;
mice: int32 := 5;`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	requireNoParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	requireStatementCount(t, program.Statements, 6)

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"cats"},
		{"dogs"},
		{"rats"},
		{"mice"},
	}

	for i, tt := range tests {
		stmt := program.Statements[i]

		if !testDeclareStatement(t, stmt, tt.expectedIdentifier) {
			return
		}
	}
}

func testDeclareStatement(t *testing.T, stmt ast.Statement, name string) bool {
	if stmt.TokenLiteral() != ":=" {
		t.Errorf("stmt.TokenLiteral not ':='. Got=%q", stmt.TokenLiteral())
		return false
	}

	declareStmt, ok := stmt.(*ast.DeclareStatement)
	if !ok {
		t.Errorf("stmt not *ast.DeclareStatement. Got=%T", stmt)
		return false
	}

	if declareStmt.Name.Value != name {
		t.Errorf("stmt.Name.Value not '%s'. Got=%s", name, declareStmt.Name.Value)
		return false
	}

	if declareStmt.Name.TokenLiteral() != name {
		t.Errorf("declareStmt.Name.TokenLiteral not '%s'. Got=%s", name, declareStmt.Name.TokenLiteral())
		return false
	}

	return true
}

func TestParseReturnStatements(t *testing.T) {
	input := `return 5;
return 10;
return 99332211;`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	requireNoParserErrors(t, p)

	requireStatementCount(t, program.Statements, 3)

	for _, stmt := range program.Statements {
		returnStatement, ok := stmt.(*ast.ReturnStatement)
		if !ok {
			t.Errorf("stmt not *ast.ReturnStatement got=%T", stmt)
			continue
		}

		if returnStatement.TokenLiteral() != "return" {
			t.Errorf("returnStatement.tokenLiteral not return. Got=%s", returnStatement.TokenLiteral())
		}
	}
}
