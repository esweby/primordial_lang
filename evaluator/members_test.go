package evaluator

import (
	"testing"

	"github.com/esweby/primordial_lang/object"
)

func TestMemberLength(t *testing.T) {
	tests := []struct {
		input  string
		output int64
	}{
		{`[]int64{20, 32, 45}.length;`, 3},
		{`[3]int64{20, 32, 45}.length;`, 3},
		{`[4]int64{20}.length;`, 4},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)

		res, ok := evaluated.(*object.Integer)
		if !ok {
			t.Fatalf("test %d failed: expected object.Integer but got %T", i, evaluated)
		}

		if res.Value != tt.output {
			t.Fatalf("test %d failed: expected %d but got %d", i, tt.output, res.Value)
		}
	}
}

func TestMemberSliceAppend(t *testing.T) {
	tests := []struct {
		input  string
		output []int64
	}{
		{`ages := []int64{20, 32, 45}; ages.append(19); ages`, []int64{20, 32, 45, 19}},
	}

	for i, tt := range tests {
		evaluated := testEval(tt.input)

		res, ok := evaluated.(*object.Slice)
		if !ok {
			t.Fatalf("test %d failed: expected object.Slice but got %T", i, evaluated)
		}

		for k, cur := range tt.output {
			testIntegerObject(t, res.Elements[k], cur)
		}
	}
}
