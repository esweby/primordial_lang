package evaluator

import "testing"

func TestFunctionsCanBeUsedBeforeTheirDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"called first", `x := f(); fn f(): int { return 1; } x;`, 1},
		{
			"mutual recursion",
			`fn even(n int): int { if (n == 0) { return 1; } return odd(n - 1); }
			fn odd(n int): int { if (n == 0) { return 0; } return even(n - 1); }
			even(10);`,
			1,
		},
		{
			"called from a method",
			`struct P { impl { fn run(): int { return later(); } } }
			fn later(): int { return 3; }
			p := P{};
			p.run();`,
			3,
		},
		{
			"called from a field default",
			`struct P { pub x = later(); }
			fn later(): int { return 4; }
			fn get(): int { p := P{}; return p.x; }
			get();`,
			4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testIntegerObject(t, testEval(t, tt.input), tt.expected)
		})
	}
}
