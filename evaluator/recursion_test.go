package evaluator

import (
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/object"
)

func TestRunawayRecursion(t *testing.T) {
	tests := []string{
		`fn f(n int): int { return f(n + 1); } f(0);`,
		`fn f(n int): int { return 1 + (if (n > 0) { [1]int{f(n + 1)}[0] } else { f(n + 1) }); } f(0);`,
	}

	for _, input := range tests {
		err, ok := testEval(t, input).(*object.Error)
		if !ok || !strings.Contains(err.Message, "stack overflow") {
			t.Errorf("%s: expected a stack overflow error, got %v", input, err)
		}
	}
}

func TestDeepRecursionWithinTheLimit(t *testing.T) {
	input := `fn sum(n int): int { if (n == 0) { return 0; } return n + sum(n - 1); } sum(50000);`
	testIntegerObject(t, testEval(t, input), 1250025000)
}
