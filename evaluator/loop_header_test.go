package evaluator

import (
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/object"
)

func TestLoopHeaderErrors(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`for (x := int8(100); x > 0; x = x + 100) { }`, "integer overflow"},
		{`for (x := int8(100); x > 0; x = x + 100) { continue; }`, "integer overflow"},
		{`mut n := 0; for (x := 1 / n; x > 0; x = x + 1) { }`, "division by zero"},
	}

	for _, tt := range tests {
		err, ok := testEval(t, tt.input).(*object.Error)
		if !ok || !strings.Contains(err.Message, tt.expected) {
			t.Errorf("%s: expected error containing %q, got %v", tt.input, tt.expected, err)
		}
	}
}
