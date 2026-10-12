package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

func TestParseFunctionLiteralWithoutParameters(t *testing.T) {
	input := `x := fn() {x := 19;}`

	function := parseFunctionLiteral(t, input)

	if len(function.Parameters) != 0 {
		t.Fatalf("expected 0 parameters. got=%d", len(function.Parameters))
	}

	if len(function.Body.Statements) != 1 {
		t.Fatalf("expected 1 function body statement. got=%d", len(function.Body.Statements))
	}
}

func TestParseFunctionLiteralWithParameters(t *testing.T) {
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

func TestParseFunctionLiteralWithReturnType(t *testing.T) {
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

func TestParseFunctionParameters(t *testing.T) {
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
				newParameter("x", types.Int32Type),
			},
		},
		{
			input: `c := fn(x int32, y int32, z int32) {}`,
			expected: expectedType{
				newParameter("x", types.Int32Type),
				newParameter("y", types.Int32Type),
				newParameter("z", types.Int32Type),
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
	requireNoParserErrors(t, p)
	requireStatementCount(t, program.Statements, 1)

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

func newParameter(paramName string, expType types.Type) *ast.Parameter {
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

func TestParseFunctionStatementWithoutParameters(t *testing.T) {
	input := `fn add() {x := 19;}`

	function := parseFunctionStatement(t, input)

	if len(function.Parameters) != 0 {
		t.Fatalf("expected 0 parameters. got=%d", len(function.Parameters))
	}

	if len(function.Body.Statements) != 1 {
		t.Fatalf("expected 1 function body statement. got=%d", len(function.Body.Statements))
	}
}

func TestParseFunctionStatementWithParameters(t *testing.T) {
	input := `fn add(x int32, y int32) {x + y;}`

	function := parseFunctionStatement(t, input)

	if len(function.Parameters) != 2 {
		t.Fatalf("expected 2 parameters. got=%d", len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0].Name, "x")
	testLiteralExpression(t, function.Parameters[1].Name, "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("expected 1 function body statement. got=%d", len(function.Body.Statements))
	}
}

func TestParseFunctionStatementWithReturnType(t *testing.T) {
	input := `fn add(x int32, y int32): int32 { x + y; }`

	function := parseFunctionStatement(t, input)

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

func parseFunctionStatement(t *testing.T, input string) *ast.FunctionStatement {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	requireNoParserErrors(t, p)
	requireStatementCount(t, program.Statements, 1)

	stmt := program.Statements[0]

	function, ok := stmt.(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionStatement. got=%T",
			stmt)
	}

	return function
}

func TestParseFunctionTypes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`f: fn(int64): int64 := g;`, "fn(int64): int64"},
		{`f: fn() := g;`, "fn(): void"},
		{`f: fn(int32, string): bool := g;`, "fn(int32, string): bool"},
		{`f: fn(int32): (int32, bool) := g;`, "fn(int32): (int32, bool)"},
		{`f: fn(fn(int8): int8): fn(): int8 := g;`, "fn(fn(int8): int8): fn(): int8"},
		{`f: fn([]int32, map[string]int64): [2]bool := g;`, "fn([]int32, map[string]int64): [2]bool"},
		{`f: function := g;`, "function"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p := New(lexer.New(tt.input))
			program := p.ParseProgram()
			requireNoParserErrors(t, p)

			stmt, ok := program.Statements[0].(*ast.DeclareStatement)
			if !ok {
				t.Fatalf("expected a declaration, got %T", program.Statements[0])
			}
			if stmt.Type.Name() != tt.want {
				t.Fatalf("type is %s, want %s", stmt.Type.Name(), tt.want)
			}
		})
	}
}

func TestParseFunctionTypesInSignatures(t *testing.T) {
	p := New(lexer.New(`fn apply(f fn(int64): int64, xs []int64): fn(int64): int64, bool { return f, true; }`))
	program := p.ParseProgram()
	requireNoParserErrors(t, p)

	fn := program.Statements[0].(*ast.FunctionStatement)
	params := []string{"fn(int64): int64", "[]int64"}
	for i, want := range params {
		if got := fn.Parameters[i].Type.Name(); got != want {
			t.Errorf("parameter %d has type %s, want %s", i, got, want)
		}
	}

	returns := []string{"fn(int64): int64", "bool"}
	if len(fn.ReturnTypes) != len(returns) {
		t.Fatalf("expected %d return types, got %d", len(returns), len(fn.ReturnTypes))
	}
	for i, want := range returns {
		if got := fn.ReturnTypes[i].Type.Name(); got != want {
			t.Errorf("return %d has type %s, want %s", i, got, want)
		}
	}
}
