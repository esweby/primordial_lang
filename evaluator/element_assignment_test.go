package evaluator

import (
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/object"
)

func TestElementAssignment(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`mut xs := []int{1, 2}; xs[0] = 5; xs[0];`, 5},
		{`mut xs := []int{1, 2}; xs[1] = xs[0] + 8; xs[1];`, 9},
		{`mut xs := [2]int{1, 2}; xs[1] = 7; xs[1];`, 7},
		{`mut xs := []int{1, 2}; xs[0] = 5; xs[1];`, 2},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(t, tt.input), tt.expected)
	}
}

func TestElementAssignmentOutOfBounds(t *testing.T) {
	for _, input := range []string{
		`mut xs := []int{1}; xs[1] = 5;`,
		`mut xs := [1]int{1}; xs[-1] = 5;`,
	} {
		err, ok := testEval(t, input).(*object.Error)
		if !ok || !strings.Contains(err.Message, "out of bounds") {
			t.Errorf("%s: expected an out of bounds error, got %v", input, err)
		}
	}
}
