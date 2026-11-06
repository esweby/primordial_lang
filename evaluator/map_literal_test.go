package evaluator

import "testing"

func TestMapLiteralDuplicateKeys(t *testing.T) {
	input := `
		m := map[string]int{"a": 1, "a": 2};
		mut visits := 0;
		mut total := 0;
		for k, v := range m {
			visits = visits + 1;
			total = total + v;
		}
		visits * 10 + total;
	`
	testIntegerObject(t, testEval(t, input), 12)
}
