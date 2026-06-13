package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
)

func TestCallExpressionParsing(t *testing.T) {
	input := `add(2, 3, 4);`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not an ExpressionStatement. Got=%T",
			program.Statements[0],
		)
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt is not a CallExpression. Got=%T",
			stmt.Expression,
		)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. Got=%d", len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 2)
	testLiteralExpression(t, exp.Arguments[1], 3)
	testLiteralExpression(t, exp.Arguments[2], 4)
}
