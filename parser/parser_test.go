package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
)

func TestDeclareStatements(t *testing.T) {
	input := `x := 5;
y := 10;
pub foobar := 548632;
mut cats := "dogs";
pub const dogs := "cats"`

	l := lexer.New(input) 
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	checkNumExpectedStatements(t, program.Statements, 5)

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"foobar"},
		{"cats"},
		{"dogs"},
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

func TestReturnStatement(t *testing.T) {
	input := `return 5;
return 10;
return 99332211;`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	checkNumExpectedStatements(t, program.Statements, 3)

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

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()

	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %s", msg)
	}

	t.FailNow()
}

func checkNumExpectedStatements(t *testing.T, stmts []ast.Statement, numExpected int) {
	if len(stmts) != numExpected {
		t.Errorf("statements does not contain %d statements. got=%d", numExpected, len(stmts))
		t.FailNow()
	}
}