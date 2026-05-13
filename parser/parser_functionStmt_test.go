package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
)

func TestFunctionStmtBasicParsing(t *testing.T) {
	input := `fn add() {x := 19;}`

	function := parseFunctionStmt(t, input)

	if len(function.Parameters) != 0 {
		t.Fatalf("expected 0 parameters. got=%d", len(function.Parameters))
	}

	if len(function.Body.Statements) != 1 {
		t.Fatalf("expected 1 function body statement. got=%d", len(function.Body.Statements))
	}
}

func TestFunctionStmtWithArguments(t *testing.T) {
	input := `fn add(x int32, y int32) {x + y;}`

	function := parseFunctionStmt(t, input)

	if len(function.Parameters) != 2 {
		t.Fatalf("expected 2 parameters. got=%d", len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0].Name, "x")
	testLiteralExpression(t, function.Parameters[1].Name, "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("expected 1 function body statement. got=%d", len(function.Body.Statements))
	}
}

func TestFunctionStmtParsing(t *testing.T) {
	input := `fn add(x int32, y int32): int32 { x + y; }`

	function := parseFunctionStmt(t, input)

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

func parseFunctionStmt(t *testing.T, input string) *ast.FunctionStatement {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt := program.Statements[0]

	function, ok := stmt.(*ast.FunctionStatement)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionStmt. got=%T",
			stmt)
	}

	return function
}
