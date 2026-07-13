package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
)

func TestStringLiterals(t *testing.T) {
	input := `"hello world";`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt := program.Statements[0].(*ast.ExpressionStatement) 
	literal, ok := stmt.Expression.(*ast.StringLiteral) 
	if !ok { 
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression) 
	} 
	
	if literal.Value != "hello world" { 
		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value) 
	}
}