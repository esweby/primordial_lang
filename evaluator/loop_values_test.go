package evaluator

import "testing"

func TestLoopsAsValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"returned", `fn f(): int { return for { break (1); }; } f() + 1;`, 2},
		{"as an operand", `(for { break (2); }) + 1;`, 3},
		{"as a struct field", `struct P { pub x: int; } p := P{x: for { break (4); }}; p.x;`, 4},
		{"as an element", `xs := []int{for { break (5); }}; xs[0];`, 5},
		{"as a break value", `for { break (for { break (6); }); };`, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testIntegerObject(t, testEval(t, tt.input), tt.expected)
		})
	}
}
