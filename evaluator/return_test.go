package evaluator

import "testing"

func TestReturnInsideLoops(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"infinite", `fn f(): int { for { return 1; } return 2; } f()`, 1},
		{"while", `fn f(): int { mut i := 0; for (i < 5) { i = i + 1; if (i == 2) { return i; } } return 99; } f()`, 2},
		{"constructed", `fn f(): int { for (i := 0; i < 5; i = i + 1) { if (i == 3) { return i; } } return 99; } f()`, 3},
		{"range", `fn f(): int { for _, x := range []int{5, 7} { if (x == 7) { return x; } } return 99; } f()`, 7},
		{"nested", `fn f(): int { for { for { return 4; } } return 99; } f()`, 4},
		{"loop as a value", `fn f(): int { x := for { return 5; break (1); }; return x + 1; } f()`, 5},
		{"if as a value", `fn f(b bool): int { x := if (b) { return 6; } else { 1 }; return x + 1; } f(true)`, 6},
		{"as an argument", `fn g(x int): int { return x + 1; } fn f(): int { return g(for { return 7; break (1); }); } f()`, 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testIntegerObject(t, testEval(t, tt.input), tt.expected)
		})
	}
}
