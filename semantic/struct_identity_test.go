package semantic

import "testing"

func TestAnalyze_StructsWithTheSameName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`struct P { pub x: int; } fn f(p P): int { return p.x; } fn g(): int { struct P { pub y: int; } return f(P{y: 1}); }`, "argument 0: expected P, got P"},
		{`struct P { pub x: int; } fn h(): P { struct P { pub y: int; } return P{y: 2}; }`, "return value 0: expected P, got P"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			errs := analyze(t, test.input)
			if !containsError(errs, test.expected) {
				t.Fatalf("expected error containing %q, got %v", test.expected, errs)
			}
		})
	}
}
