package semantic

import "testing"

func TestAnalyze_VoidValues(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`fn f() {} x := f();`, "cannot declare x from a call that returns nothing"},
		{`fn f() {} fn g() { struct P { x = f(); } }`, "default value for P.x is a call that returns nothing"},
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
