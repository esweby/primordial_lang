package semantic

import "testing"

func TestAnalyze_MapKeyTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`struct P { pub x: int; } m := map[P]int{};`, "type P cannot be a map key"},
		{`m := map[function]int{};`, "type function cannot be a map key"},
		{`m := map[float64]int{};`, "type float64 cannot be a map key"},
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
