package evaluator

import "testing"

func TestConstantExpressionsFoldBeforeTheyAreTyped(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`x: int8 := 100 + 100 - 100; x;`, 100},
		{`x: uint8 := 1 - 2 + 3; x;`, 2},
		{`x: uint := -1 + 2; x;`, 1},
		{`x: int8 := -128; x;`, -128},
		{`x: int8 := 1; x + (100 + 100 - 100);`, 101},
		{`fn f(x int8): int8 { return x; } f(100 + 100 - 100);`, 100},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(t, tt.input), tt.expected)
	}
}
