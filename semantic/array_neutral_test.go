package semantic

import "testing"

func TestAnalyze_ArraysWithoutNeutralValues(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`x := [2]float64{};`, "array of float64 has no neutral value"},
		{`x := [2]function{};`, "array of function has no neutral value"},
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
