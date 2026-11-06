package evaluator

import "testing"

func TestStructsCanBeUsedBeforeTheirDeclaration(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`p := P{x: 1}; struct P { pub x: int; } p.x;`, 1},
		{`fn f(): int { p := P{x: 2}; return p.x; } v := f(); struct P { pub x: int; } v;`, 2},
		{`v := P.make(); struct P { pub x: int; fn make(): int { return 3; } } v;`, 3},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(t, tt.input), tt.expected)
	}
}
