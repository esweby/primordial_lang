package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
)

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x; }`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()

	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.statements[0] is not ast.ExpressionStatement. Got=%T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.IfExpression)
	if !ok {
		t.Fatalf("stmt.Expression not ast.IfExpression. Got=%T", stmt.Expression)
	}

	if !testInfixExpression(t, exp.Condition, "x", "<", "y") {
		return
	}

	checkNumExpectedStatements(t, exp.Body.Statements, 1)

	body, ok := exp.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.statements[0] is not ast.ExpressionStatement. Got=%T", exp.Body.Statements[0])
	}

	if !testIdentifier(t, body.Expression, "x") {
		return
	}

	if exp.Else != nil {
		t.Errorf("exp.Else not nil. Got=%+v", exp.Else)
	}
}

func TestIfElseExpression(t *testing.T) {
	input := `if (x < y) { x; } else { z; }`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()

	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.statements[0] is not ast.ExpressionStatement. Got=%T", program.Statements[0])
	}

	exp, _ := stmt.Expression.(*ast.IfExpression)

	if exp.Else == nil {
		t.Errorf("exp.Else is nil")
	}

	elseBody, ok := exp.Else.(*ast.BlockExpression)
	if !ok {
		t.Fatalf("exp.Else is not ast.BlockExpression. Got=%T", exp.Else)
	}

	stmt = elseBody.Statements[0].(*ast.ExpressionStatement)
	if !testIdentifier(t, stmt.Expression, "z") {
		return
	}
}

func TestIfElseIfExpression(t *testing.T) {
	input := `if (x < y) { x; } else if (x > y) { z; }`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()

	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.statements[0] is not ast.ExpressionStatement. Got=%T", program.Statements[0])
	}

	exp, _ := stmt.Expression.(*ast.IfExpression)

	if exp.Else == nil {
		t.Errorf("exp.Else is nil")
	}

	el, ok := exp.Else.(*ast.IfExpression)
	if !ok {
		t.Fatalf("exp.Else is not ast.IfExpression. Got=%T", exp.Else)
	}

	if !testInfixExpression(t, el.Condition, "x", ">", "y") {
		return
	}

	checkNumExpectedStatements(t, el.Body.Statements, 1)

	body, ok := el.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.statements[0] is not ast.ExpressionStatement. Got=%T", exp.Body.Statements[0])
	}

	if !testIdentifier(t, body.Expression, "z") {
		return
	}

	if el.Else != nil {
		t.Errorf("exp.Exp not nil. Got=%+v", exp.Else)
	}
}

func TestIfExpressionInDeclaration(t *testing.T) {
	input := `a := if (x < y) { x; } else { z; };`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()

	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	if !testDeclareStatement(t, program.Statements[0], "a") {
		return
	}

	stmt, ok := program.Statements[0].(*ast.DeclareStatement)
	if !ok {
		t.Fatalf("program.statements[0] is not ast.DeclareStatement. Got=%T", program.Statements[0])
	}

	ifExpr, ok := stmt.Value.(*ast.IfExpression)
	if !ok {
		t.Fatalf("program.statements[0] is not ast.IfExpression. Got=%T", program.Statements[0])
	}

	if !testInfixExpression(t, ifExpr.Condition, "x", "<", "y") {
		return
	}

	checkNumExpectedStatements(t, ifExpr.Body.Statements, 1)

	body, ok := ifExpr.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.statements[0] is not ast.ExpressionStatement. Got=%T", ifExpr.Body.Statements[0])
	}

	if !testIdentifier(t, body.Expression, "x") {
		return
	}

	if ifExpr.Else == nil {
		t.Errorf("exp.Else is nil")
	}

	elseBody, ok := ifExpr.Else.(*ast.BlockExpression)
	if !ok {
		t.Fatalf("exp.Else is not ast.BlockExpression. Got=%T", ifExpr.Else)
	}

	elseStmt, _ := elseBody.Statements[0].(*ast.ExpressionStatement)

	if !testIdentifier(t, elseStmt.Expression, "z") {
		return
	}
}