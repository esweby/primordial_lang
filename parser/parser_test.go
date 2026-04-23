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
mut cats := 12;`

	l := lexer.New(input) 
	p := New(l)

	program := p.ParseProgram()
	checkParserErrors(t, p)

	if program == nil {
		t.Fatalf("ParseProgram() returned nil")
	}

	checkNumExpectedStatements(t, program.Statements, 3)

	tests := []struct {
		expectedIdentifier string
	}{
		{"x"},
		{"y"},
		{"cats"},
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

func TestBooleanExpression(t *testing.T) {
	input := `true;`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram()

	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not an ast.ExpressionStatement. Got=%T", program.Statements[0])
	}

	boolean, ok := stmt.Expression.(*ast.Boolean)
	if !ok {
		t.Fatalf("exp not *ast.Identifier. got=%T", stmt.Expression)
	}

	if boolean.Value != true {
		t.Errorf("ident.Value not %s. Got=%s", "true", boolean.TokenLiteral())
	}

	if boolean.TokenLiteral() != "true" {
		t.Errorf("ident.TokenLiteral() is not %s. Got=%s", "true", boolean.TokenLiteral())
	}
}

func TestIfExpression(t *testing.T) {
	input := `if (x < y) { x; }`

	l := lexer.New(input)
	p := New(l)

	program := p.ParseProgram();

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

	program := p.ParseProgram();

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

	program := p.ParseProgram();

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

	program := p.ParseProgram();

	checkParserErrors(t, p)
	checkNumExpectedStatements(t, program.Statements, 1)

	if !testDeclareStatement(t, program.Statements[0], "a") {
		return
	}

	// stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	// if !ok {
	// 	t.Fatalf("program.statements[0] is not ast.ExpressionStatement. Got=%T", program.Statements[0])
	// }

	// exp, _ := stmt.Expression.(*ast.IfExpression)

	// if exp.Else == nil {
	// 	t.Errorf("exp.Else is nil")
	// }

	// el, ok := exp.Else.(*ast.IfExpression)
	// if !ok {
	// 	t.Fatalf("exp.Else is not ast.IfExpression. Got=%T", exp.Else)
	// }

	// if !testInfixExpression(t, el.Condition, "x", "<", "y") {
	// 	return
	// }

	// checkNumExpectedStatements(t, el.Body.Statements, 1)

	// body, ok := el.Body.Statements[0].(*ast.ExpressionStatement)
	// if !ok {
	// 	t.Fatalf("program.statements[0] is not ast.ExpressionStatement. Got=%T", exp.Body.Statements[0])
	// }

	// if !testIdentifier(t, body.Expression, "z") {
	// 	return
	// }

	// if el.Else != nil {
	// 	t.Errorf("exp.Exp not nil. Got=%+v", exp.Else)
	// }
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct{
		input string
		operator string
		value interface{}
	}{
		{"!5;", "!", 5},
		{"-15", "-", 15},
		{"!true","!",true},
		{"!false","!",false},
	}

	for _, tt := range prefixTests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		checkNumExpectedStatements(t, program.Statements, 1)

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
					program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("stmt is not ask.PrefixExpression. got=%T", stmt.Expression)
		}

		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s",
				tt.operator, exp.Operator)
		}
		if !testLiteralExpression(t, exp.Right, tt.value) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct{
		input string
		leftValue interface{}
		operator string
		rightValue interface{}
	}{
		{"5 + 5;", 5, "+", 5},
		{"5 - 5;", 5, "-", 5},
		{"5 * 5;", 5, "*", 5},
		{"5 / 5;", 5, "/", 5},
		{"5 > 5;", 5, ">", 5},
		{"5 < 5;", 5, "<", 5},
		{"5 == 5;", 5, "==", 5},
		{"5 != 5;", 5, "!=", 5},
		{"true == true;", true, "==", true},
		{"true != false;", true, "!=", false},
		{"false == false;", false, "==", false},
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

		if testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue) {
			return 
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input string
		expected string
	}{
		{"-a * b","((-a) * b)"},
		{"!-a","(!(-a))"},
		{"a + b + c","((a + b) + c)"},
		{"a * b * c","((a * b) * c)"},
		{"a * b / c","((a * b) / c)"},
		{"a + b * c + d / e - f","(((a + (b * c)) + (d / e)) - f)"},
		{"3 + 4;-5 * 5","(3 + 4)((-5) * 5)"},
		{"5 > 4 == 3 < 4","((5 > 4) == (3 < 4))"},
		{"5 < 4 != 3 > 4","((5 < 4) != (3 > 4))"},
		{"3 + 4 * 5 == 3 * 1 + 4 * 5","((3 + (4 * 5)) == ((3 * 1) + (4 * 5)))"},
		{"true", "true"},
		{"false", "false"},
		{"3 > 5 == false", "((3 > 5) == false)"},
		{"3 < 5 == true", "((3 < 5) == true)"},
		{"1 + (2 + 3) + 4","((1 + (2 + 3)) + 4)"},
		{"(5 + 5) * 2","((5 + 5) * 2)"},
		{"2 / (5 + 5)","(2 / (5 + 5))"},
		{"!(true == true)","(!(true == true))"},
		// {"a + add(b * c) + d", "((a + add((b * c))) + d)"},
		// {"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))", "add(a,b,1,(2 * 3),(4 + 5),add(6,(7 * 8)))"},
		// {"add(a + b + c * d / f + g)", "add((((a + b) + ((c * d) / f)) + g))"},
		// {"a * [1, 2, 3, 4][b * c] * d","((a * ([1,2,3,4][(b * c)])) * d)"},
		// {"add(a * b[2], b[1], 2 * [1, 2][1])", "add((a * (b[2])),(b[1]),(2 * ([1,2][1])))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t,p)

		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
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

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not ast.Identifier. Got=%T", exp)
		return false
	}

	if ident.Value != value {
		t.Errorf("ident.Value not =%s. Got=%s", value, ident.Value)
		return false
	}

	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not =%s. Got=%s", value, ident.TokenLiteral())
		return false
	}

	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	bo, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}

	if bo.Value != value {
		t.Errorf("bo.Value not %t. got=%t", value, bo.Value)
		return false
	}

	if bo.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("bo.TokenLiteral not %t. got=%s",
			value, bo.TokenLiteral())
		return false
	}

	return true
}

func testLiteralExpression(t *testing.T, exp ast.Expression, expected interface{}) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}

	t.Errorf("type of exp not handled. Got=%T", exp)
	return false
}

func testInfixExpression(
	t *testing.T, 
	exp ast.Expression, 
	left interface{}, 
	operator string,
	right interface{},
) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp not *ast.InfixExpression. got=%T(%s)", exp,exp)
		return false
	}

	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}

	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}

	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}

	return true
}
