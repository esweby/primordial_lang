package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

func TestFunctionLiteralBasicParsing(t *testing.T) {
	input := `x := fn() {x := 19;}`

	function := parseFunctionLiteral(t, input)

	if len(function.Parameters) != 0 {
		t.Fatalf("expected 0 parameters. got=%d", len(function.Parameters))
	}

	if len(function.Body.Statements) != 1 {
		t.Fatalf("expected 1 function body statement. got=%d", len(function.Body.Statements))
	}
}

func TestFunctionLiteralWithArguments(t *testing.T) {
	input := `x := fn(x int32, y int32) { x + y; }`

	function := parseFunctionLiteral(t, input)

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
	input := `x := fn(x int32, y int32): int32 { x + y; }`

	function := parseFunctionLiteral(t, input)

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
			input:    `a := fn() {}`,
			expected: expectedType{},
		},
		{
			input: `b := fn(x int32) {}`,
			expected: expectedType{
				createParameterToken("x", &types.Int32{}),
			},
		},
		{
			input: `c := fn(x int32, y int32, z int32) {}`,
			expected: expectedType{
				createParameterToken("x",  &types.Int32{}),
				createParameterToken("y",  &types.Int32{}),
				createParameterToken("z",  &types.Int32{}),
			},
		},
	}

	for _, tt := range tests {
		function := parseFunctionLiteral(t, tt.input)

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

func parseFunctionLiteral(t *testing.T, input string) *ast.FunctionLiteral {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt := program.Statements[0]

	declareStmt, ok := stmt.(*ast.DeclareStatement)
	if !ok {
		t.Errorf("stmt not *ast.DeclareStatement. Got=%T", stmt)
	}

	function, ok := declareStmt.Value.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T",
			declareStmt.Value)
	}

	return function
}

func createParameterToken(paramName string, expType types.Type) *ast.Parameter {
	return &ast.Parameter{
		Name: &ast.Identifier{
			Token: token.Token{
				Type:    token.STRING_LITERAL,
				Literal: paramName,
			},
			Value: paramName,
		},
		Type: expType,
	}
}
