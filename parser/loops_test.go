package parser

import (
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
)

func TestBasicLoop(t *testing.T) {
	tests := []struct {
		input                  string
		expectedBodyStatements int
		testFunc               func(
			t *testing.T,
			forLoop *ast.ForLoop,
			testNum int,
		) (ast.ForController, bool)
	}{
		{`for { x := 1; }`, 1, testIsInfiniteLoop},
		{`for {}`, 0, testIsInfiniteLoop},
		{`for x := range y {}`, 0, testIsRangeLoop},
		{`for x := range map[int32]string{ 1: "one", } { x + y; } `, 1, testIsRangeLoop},
		{`for i, y := range [3]int32{1, 2, 3} { return i * y; }`, 1, testIsRangeLoop},
		{`for (x < y) {}`, 0, testIsWhileLoop},
		{`for (x := 1; x < 10; x = x + 1) {}`, 0, testIsConstructedLoop},
	}

	for i, tt := range tests {
		forExp := testExpressionStatementLoop(t, tt.input, i)

		_, ok := tt.testFunc(t, forExp, i)
		if !ok {
			return
		}

		body := forExp.Body
		if len(body.Statements) != tt.expectedBodyStatements {
			t.Errorf(
				"test %d: expected body with %d statements, got %d",
				i,
				tt.expectedBodyStatements,
				len(body.Statements),
			)
		}
	}
}

func TestLoopInDeclare(t *testing.T) {
	tests := []struct {
		input                  string
		expectedBodyStatements int
		testFunc               func(
			t *testing.T,
			forLoop *ast.ForLoop,
			testNum int,
		) (ast.ForController, bool)
	}{
		{`x := for { x := 1; }`, 1, testIsInfiniteLoop},
		{`y := for {}`, 0, testIsInfiniteLoop},
		{`z := for a := range cats {}`, 0, testIsRangeLoop},
	}

	for i, tt := range tests {
		forExp := testDeclareExpressionLoop(t, tt.input, i)

		_, ok := tt.testFunc(t, forExp, i)
		if !ok {
			return
		}

		body := forExp.Body
		if len(body.Statements) != tt.expectedBodyStatements {
			t.Errorf(
				"test %d: expected body with %d statements, got %d",
				i,
				tt.expectedBodyStatements,
				len(body.Statements),
			)
		}
	}
}

func TestLoopWithLabels(t *testing.T) {
	tests := []struct {
		input         string
		expectedLabel string
		testFunc      func(
			t *testing.T,
			forLoop *ast.ForLoop,
			testNum int,
		) (ast.ForController, bool)
	}{
		{`lbl: for {}`, "lbl", testIsInfiniteLoop},
		{`lbl: for (x < y) {}`, "lbl", testIsWhileLoop},
		{`lbl: for (x := 1; x < y; x = x + 1) {}`, "lbl", testIsConstructedLoop},
		{`lbl: for x := range brian {}`, "lbl", testIsRangeLoop},
	}

	for i, tt := range tests {
		forExp := testExpressionStatementLoop(t, tt.input, i)

		_, ok := tt.testFunc(t, forExp, i)
		if !ok {
			return
		}

		if forExp.Label.Value != tt.expectedLabel {
			t.Errorf(
				"test %d: expected for label to be %s got %s",
				i,
				tt.expectedLabel,
				forExp.Label.Value,
			)
		}
	}
}

func TestBreakStatement(t *testing.T) {
	tests := []struct {
		input    string
		testFunc func(
			t *testing.T,
			forLoop *ast.ForLoop,
			testNum int,
		) (ast.ForController, bool)
	}{
		{`for { break; }`, testIsInfiniteLoop},
		{`for (x < y) { break; }`, testIsWhileLoop},
		{`for (x := 1; x < y; x = x + 1) { break; }`, testIsConstructedLoop},
		{`for x := range brian { break; }`, testIsRangeLoop},
	}

	for i, tt := range tests {
		forExp := testExpressionStatementLoop(t, tt.input, i)

		_, ok := tt.testFunc(t, forExp, i)
		if !ok {
			return
		}

		stmt := forExp.Body.Statements[0]

		_, ok = stmt.(*ast.BreakStatement)
		if !ok {
			t.Errorf(
				"test %d: expected body to be ast.BreakStatement, got %T",
				i,
				forExp.Body.Statements[0],
			)
		}
	}
}

func TestBreakWithLabels(t *testing.T) {
	tests := []struct {
		input              string
		expectedLabel      string
		expectedBreakLabel string
		testFunc           func(
			t *testing.T,
			forLoop *ast.ForLoop,
			testNum int,
		) (ast.ForController, bool)
	}{
		{`lbl: for { break lbl; }`, "lbl", "lbl", testIsInfiniteLoop},
		{`lbl: for (x < y) { break lbl; }`, "lbl", "lbl", testIsWhileLoop},
		{`lbl: for (x := 1; x < y; x = x + 1) { break lbl; }`, "lbl", "lbl", testIsConstructedLoop},
		{`lbl: for x := range brian { break lbl; }`, "lbl", "lbl", testIsRangeLoop},
	}

	for i, tt := range tests {
		forExp := testExpressionStatementLoop(t, tt.input, i)

		_, ok := tt.testFunc(t, forExp, i)
		if !ok {
			return
		}

		if forExp.Label.Value != tt.expectedLabel {
			t.Errorf(
				"test %d: expected for label to be %s got %s",
				i,
				tt.expectedLabel,
				forExp.Label.Value,
			)
		}

		stmt := forExp.Body.Statements[0]

		brkStmt, ok := stmt.(*ast.BreakStatement)
		if !ok {
			t.Errorf(
				"test %d: expected body to be ast.BreakStatement, got %T",
				i,
				forExp.Body.Statements[0],
			)
		}

		if brkStmt.Label.Value != tt.expectedLabel {
			t.Errorf(
				"test %d: expected for label to be %s got %s",
				i,
				tt.expectedLabel,
				brkStmt.Label.Value,
			)
		}
	}
}

func TestBreakWithExpressions(t *testing.T) {
	tests := []struct {
		input              string
		testFunc           func(
			t *testing.T,
			forLoop *ast.ForLoop,
			testNum int,
		) (ast.ForController, bool)
	}{
		{`lbl: for { break lbl ("yay"); }`, testIsInfiniteLoop},
		{`lbl: for (x < y) { break lbl (1 + 1); }`, testIsWhileLoop},
		{`lbl: for (x := 1; x < y; x = x + 1) { break lbl (1 + 2); }`, testIsConstructedLoop},
		{`lbl: for x := range brian { break lbl (4); }`, testIsRangeLoop},
	}

	for i, tt := range tests {
		forExp := testExpressionStatementLoop(t, tt.input, i)

		_, ok := tt.testFunc(t, forExp, i)
		if !ok {
			return
		}

		stmt := forExp.Body.Statements[0]

		brkStmt, ok := stmt.(*ast.BreakStatement)
		if !ok {
			t.Errorf(
				"test %d: expected body to be ast.BreakStatement, got %T",
				i,
				forExp.Body.Statements[0],
			)
		}

		if brkStmt.Value == nil {
			t.Errorf(
				"test %d: expected brkStmt.Value to not be nil",
				i,
			)
		}
	}
}

func TestBreakWithLabelsAndExpressions(t *testing.T) {
	tests := []struct {
		input              string
		expectedLabel      string
		expectedBreakLabel string
		testFunc           func(
			t *testing.T,
			forLoop *ast.ForLoop,
			testNum int,
		) (ast.ForController, bool)
	}{
		{`lbl: for { break lbl ("yay"); }`, "lbl", "lbl", testIsInfiniteLoop},
		{`lbl: for (x < y) { break lbl (1 + 1); }`, "lbl", "lbl", testIsWhileLoop},
		{`lbl: for (x := 1; x < y; x = x + 1) { break lbl (1 + 2); }`, "lbl", "lbl", testIsConstructedLoop},
		{`lbl: for x := range brian { break lbl (4); }`, "lbl", "lbl", testIsRangeLoop},
	}

	for i, tt := range tests {
		forExp := testExpressionStatementLoop(t, tt.input, i)

		_, ok := tt.testFunc(t, forExp, i)
		if !ok {
			return
		}

		if forExp.Label.Value != tt.expectedLabel {
			t.Errorf(
				"test %d: expected for label to be %s got %s",
				i,
				tt.expectedLabel,
				forExp.Label.Value,
			)
		}

		stmt := forExp.Body.Statements[0]

		brkStmt, ok := stmt.(*ast.BreakStatement)
		if !ok {
			t.Errorf(
				"test %d: expected body to be ast.BreakStatement, got %T",
				i,
				forExp.Body.Statements[0],
			)
		}

		if brkStmt.Label.Value != tt.expectedLabel {
			t.Errorf(
				"test %d: expected brkStmt.label to be %s got %s",
				i,
				tt.expectedLabel,
				brkStmt.Label.Value,
			)
		}

		if brkStmt.Value == nil {
			t.Errorf(
				"test %d: expected brkStmt.Value to not be nil",
				i,
			)
		}
	}
}

func TestContinueStatement(t *testing.T) {
	tests := []struct {
		input    string
		testFunc func(
			t *testing.T,
			forLoop *ast.ForLoop,
			testNum int,
		) (ast.ForController, bool)
	}{
		{`for { continue; }`, testIsInfiniteLoop},
		{`for (x < y) { continue; }`, testIsWhileLoop},
		{`for (x := 1; x < y; x = x + 1) { continue; }`, testIsConstructedLoop},
		{`for x := range brian { continue; }`, testIsRangeLoop},
	}

	for i, tt := range tests {
		forExp := testExpressionStatementLoop(t, tt.input, i)

		_, ok := tt.testFunc(t, forExp, i)
		if !ok {
			return
		}

		stmt := forExp.Body.Statements[0]

		_, ok = stmt.(*ast.ContinueStatement)
		if !ok {
			t.Errorf(
				"test %d: expected body to be ast.ContinueStatement, got %T",
				i,
				forExp.Body.Statements[0],
			)
		}
	}
}

func testExpressionStatementLoop(t *testing.T, input string, testNum int) *ast.ForLoop {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	requireNoParserErrors(t, p)
	requireStatementCount(t, program.Statements, 1)

	stmt := program.Statements[0].(*ast.ExpressionStatement)

	forExpression, ok := stmt.Expression.(*ast.ForLoop)
	if !ok {
		t.Fatalf("test %d: stmt.Expression not *ast.ForLoop, got=%T", testNum, stmt.Expression)
	}

	return forExpression
}

func testDeclareExpressionLoop(t *testing.T, input string, testNum int) *ast.ForLoop {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	requireNoParserErrors(t, p)
	requireStatementCount(t, program.Statements, 1)

	stmt, ok := program.Statements[0].(*ast.DeclareStatement)

	forExpression, ok := stmt.Value.(*ast.ForLoop)
	if !ok {
		t.Fatalf("test %d: stmt.Value not *ast.ForLoop, got=%T", testNum, stmt.Value)
	}

	return forExpression
}

func testIsInfiniteLoop(t *testing.T, forLoop *ast.ForLoop, testNum int) (ast.ForController, bool) {
	infiniteLoop, ok := forLoop.Controller.(*ast.Infinite)
	if !ok {
		t.Errorf("test %d: forLoop.Controller is not *ast.Infinite, got=%T", testNum, forLoop.Controller)
		return nil, false
	}

	return infiniteLoop, true
}

func testIsRangeLoop(t *testing.T, forLoop *ast.ForLoop, testNum int) (ast.ForController, bool) {
	rangeLoop, ok := forLoop.Controller.(*ast.Range)
	if !ok {
		t.Errorf("test %d: forLoop.Controller is not *ast.Range, got=%T", testNum, forLoop.Controller)
		return nil, false
	}

	return rangeLoop, true
}

func testIsWhileLoop(t *testing.T, forLoop *ast.ForLoop, testNum int) (ast.ForController, bool) {
	rangeLoop, ok := forLoop.Controller.(*ast.While)
	if !ok {
		t.Errorf("test %d: forLoop.Controller is not *ast.While, got=%T", testNum, forLoop.Controller)
		return nil, false
	}

	return rangeLoop, true
}

func testIsConstructedLoop(t *testing.T, forLoop *ast.ForLoop, testNum int) (ast.ForController, bool) {
	rangeLoop, ok := forLoop.Controller.(*ast.Constructed)
	if !ok {
		t.Errorf("test %d: forLoop.Controller is not *ast.Constructed, got=%T", testNum, forLoop.Controller)
		return nil, false
	}

	return rangeLoop, true
}
