package evaluator

import (
	"testing"

	"github.com/esweby/primordial_lang/object"
)

func TestLenBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("five")`, 4},
		{`len([3]int64{1, 2, 3})`, 3},
		{`len([]int64{1, 2, 3, 4, 5})`, 5},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Fatalf("test %d: expected error obj. Got=%T (%+v)", i, evaluated, evaluated)
				continue
			}

			if errObj.Message != tt.expected {
				t.Fatalf("test %d: expected msg (%s) got=%s", i, tt.expected, errObj.Message)
			}
		}
	}
}
