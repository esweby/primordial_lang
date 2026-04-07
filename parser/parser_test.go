package parser

import (
	"fmt"
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
)

func TestDeclareStatements(t *testing.T) {
	input := `x := 5;
y := 10;
pub foobar := "548632";
mut cats := 12;
pub const dogs := "Dogs";`

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

func TestIdentifierExpression(t *testing.T) {
	testWord := "foobar"
	input := "foobar"

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()

	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. Got=%T", program.Statements[0])
	}

	ident, ok := stmt.Expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}

	if ident.Value != testWord {
		t.Errorf("ident.Value not %s. Got=%s", testWord, ident.Value)
	}

	if ident.TokenLiteral() != testWord {
		t.Errorf("ident.TokenLiteral() is not %s. Got=%s", testWord, ident.TokenLiteral())
	}
}

func TestIntegerLiteralExpressions(t *testing.T) {
	input := `5;`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()

	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. Got=%T", program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}

	if literal.Value != 5 {
		t.Errorf("literal.Value not %d. Got=%d", 5, literal.Value)
	}

	if literal.TokenLiteral() != "5" {
		t.Errorf("literal.TokenLiteral() is not %d. Got=%s", 5, literal.TokenLiteral())
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct{
		input string
		operator string
		integerValue int64
	}{
		{"!5;", "!", 5},
		{"-15", "-", 15},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		checkNumExpectedStatements(t, program.Statements, 1)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] not *ast.ExpressionStatement. Got=%T",
				program.Statements[0],
			)
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt.Expression not ast.PrefixExpression. Got=%T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator not %s. Got=%s", tt.operator, exp.Operator)
		}

		if !testIntegerLiteral(t, exp.Right, tt.integerValue) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct{
		input string
		leftValue int64
		operator string
		rightValue int64
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
	}

	for _, tt := range infixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		checkNumExpectedStatements(t, program.Statements, 1)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] not *ast.ExpressionStatement. Got=%T",
				program.Statements[0],
			)
		}

		exp, ok := stmt.Expression.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("stmt.Expression not ast.InfixExpression. Got=%T", stmt.Expression)
		}

		if !testIntegerLiteral(t, exp.Left, tt.leftValue) {
			return
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator not %s. Got=%s", tt.operator, exp.Operator)
		}

		if !testIntegerLiteral(t, exp.Right, tt.rightValue) {
			return
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
		for _, stmt := range stmts {
			t.Errorf("%t", stmt)
		}
		t.FailNow()
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	inte, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not ast.IntegerLiteral. Got=%T", inte)
		return false
	}

	if inte.Value != value {
		t.Errorf("inte.Value not %d. Got=%d", value, inte.Value)
		return false
	}

	if inte.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("inte.TokenLiteral not %d. Got=%s", value, inte.TokenLiteral())
		return false
	}

	return true
}