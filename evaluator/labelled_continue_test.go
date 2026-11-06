package evaluator

import "testing"

func TestLabelledContinue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{
			"skips the rest of the outer loop's body",
			`mut n := 0;
			outer: for (i := 0; i < 3; i = i + 1) {
				for (j := 0; j < 3; j = j + 1) {
					n = n + 1;
					continue outer;
				}
				n = n + 100;
			}
			n;`,
			3,
		},
		{
			"unlabelled stays in the inner loop",
			`mut n := 0;
			outer: for (i := 0; i < 2; i = i + 1) {
				for (j := 0; j < 2; j = j + 1) {
					continue;
				}
				n = n + 1;
			}
			n;`,
			2,
		},
		{
			"own label",
			`mut n := 0;
			outer: for (i := 0; i < 2; i = i + 1) {
				n = n + 1;
				continue outer;
			}
			n;`,
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testIntegerObject(t, testEval(t, tt.input), tt.expected)
		})
	}
}
