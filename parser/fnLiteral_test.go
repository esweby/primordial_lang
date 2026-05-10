package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/token"
)

func TestFunctionLiteralBasicParsing(t *testing.T) {
	input := `fn() {x := 19;}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T",
			stmt.Expression)
	}

	if len(function.Parameters) != 0 {
		t.Fatalf("expected 0 parameters. got=%d", len(function.Parameters))
	}

	if len(function.Body.Statements) != 1 {
		t.Fatalf("expected 1 function body statement. got=%d", len(function.Body.Statements))
	}
}

func TestFunctionLiteralWithArguments(t *testing.T) {
	input := `fn(x int32, y int32) {x + y;}`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T",
			stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("expected 2 parameters. got=%d", len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0].Name, "x")
	testLiteralExpression(t, function.Parameters[1].Name, "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("expected 1 function body statement. got=%d", len(function.Body.Statements))
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `fn(x int32, y int32): int32 { x + y; }`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T",
			stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("expected 2 parameters. got=%d", len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0].Name, "x")
	testLiteralExpression(t, function.Parameters[1].Name, "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("expected 1 function body statement. got=%d", len(function.Body.Statements))
	}

	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function.Body.Statements[0] is not ast.ExpressionStatement. got=%T",
			function.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestParsingParameters(t *testing.T) {
	type expectedType = []*ast.Parameter
	tests := []struct {
		input    string
		expected expectedType
	}{
		{
			input:    `fn() {}`,
			expected: expectedType{},
		},
		{
			input: `fn(x int32) {}`,
			expected: expectedType{
				createParameterToken("x", "int32"),
			},
		},
		{
			input: `fn(x int32, y int32, z int32) {}`,
			expected: expectedType{
				createParameterToken("x", "int32"),
				createParameterToken("y", "int32"),
				createParameterToken("z", "int32"),
			},
		},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		function := stmt.Expression.(*ast.FunctionLiteral)

		if len(function.Parameters) != len(tt.expected) {
			t.Errorf(
				"length parameters wrong. expected=%d got=%d",
				len(tt.expected),
				len(function.Parameters),
			)
		}

		for i, ident := range tt.expected {
			if function.Parameters[i].Name.Value != ident.Name.Value {
				t.Errorf(
					"parameter name wrong. expected=%s got=%s",
					ident.Name.Value,
					function.Parameters[i].Name.Value,
				)
			}
		}
	}
}

func createParameterToken(paramName string, typeName string) *ast.Parameter {
	return &ast.Parameter{
		Name: &ast.Identifier{
			Token: token.Token{
				Type:    token.STRING_LITERAL,
				Literal: paramName,
			},
			Value: paramName,
		},
		Type: typeName,
	}
}
