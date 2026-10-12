package evaluator

import (
	"testing"

)

func TestDeclareStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`a := 5; a;`, 5},
		{`a := 5 * 5; a;`, 25},
		{`a := 5; b := a; b;`, 5},
		{`a := 5; b := 5; c := a * b; c;`, 25},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(t, tt.input), tt.expected)
	}
}

func TestReturnIntStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"fn f(): int { return 10; } f()", 10},
		{"fn f(): int { return 10; 9; } f()", 10},
		{"fn f(): int { 9; return 10; } f()", 10},
		{"fn f(): int { 9; return 10 * 2; 10 } f()", 20},
		{`fn f(): int { if (10 > 1) { if (10 > 1) { return 1; } } return 0; } f()`, 1},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(t, tt.input), tt.expected)
	}
}
func TestFunctionStatement(t *testing.T) {
	tests := []struct {
		input     string
		numParams int
		numReturn int
	}{
		{`fn add(): int32 { return 3 + 3; } add`, 0, 1},
		{`fn add(x int32, y int32): int32 { return x + y; } add;`, 2, 1},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		testFunction(t, evaluated, i, tt.numParams, tt.numReturn)
	}
}

func TestTupleDeclaration(t *testing.T) {
	evaluated := testEval(t, `
		fn values(): int32, bool { return 10, true; }
		(number, _) := values();
		number;
	`)

	testIntegerObject(t, evaluated, 10)
}

func TestTupleAssignment(t *testing.T) {
	evaluated := testEval(t, `
		fn values(): int32, int32 { return 10, 20; }
		mut first: int32 := 0;
		mut second: int32 := 0;
		(first, second) = values();
		first + second;
	`)

	testIntegerObject(t, evaluated, 30)
}
