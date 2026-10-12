package evaluator

import (
	"fmt"
	"testing"

	"github.com/esweby/primordial_lang/object"
)

func TestFunctionLiteral(t *testing.T) {
	tests := []struct {
		input     string
		numParams int
		numReturn int
	}{
		{`(fn(x int32, y int32): int32 { return x + y; });`, 2, 1},
		{`add := fn(x int32, y int32): int32 { return x + y; }; add;`, 2, 1},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		testFunction(t, evaluated, i, tt.numParams, tt.numReturn)
	}
}

func TestEvalIntegerExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 - 50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true}, //0
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"2 > 1", true},
		{"1 == 1", true}, // 7
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"true == false", false},
		{"false == false", true},
		{"false == true", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 < 2) != true", false},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		testBooleanObject(t, i, evaluated, tt.expected)
	}
}

func TestIfElseExpr(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 > 2) { 10 } else if(2 > 1) { 20 }", 20},
		{"if (1 > 2) { 10 } else if(2 == 1) { 20 } else { 30 }", 30},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestArrayOperatorExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{`[3]int32{0,1,2}[0]`, 0},
		{`[3]int32{0,1,2}[1]`, 1},
		{`[3]int32{0,1,2}[2]`, 2},
		{`[3]int32{0,1,2}[1+1]`, 2},
		{`[3]int32{0,1,2}[2-1]`, 1},
		{`[3]int32{0,1,2}[3-1]`, 2},
		{`[3]int32{0,1,2}[-1]`, nil},
		{`[3]int32{0,1,2}[3]`, nil},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testErrorObject(t, evaluated)
		}
	}
}

func TestSliceOperatorExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{`[]int32{0,1,2}[0]`, 0},
		{`[]int32{0,1,2}[1]`, 1},
		{`[]int32{0,1,2}[2]`, 2},
		{`[]int32{0,1,2}[1+1]`, 2},
		{`[]int32{0,1,2}[2-1]`, 1},
		{`[]int32{0,1,2}[3-1]`, 2},
		{`[]int32{0,1,2}[-1]`, nil},
		{`[]int32{0,1,2}[3]`, nil},
	}

	for _, tt := range tests {
		evaluated := testEval(t, tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testErrorObject(t, evaluated)
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected []int64
	}{
		{`[3]int32{1, 2, 3}`, []int64{1, 2, 3}},
		{`[3]int64{1}`, []int64{1, 0, 0}},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		result, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("test %d: expected object.Array, got=%T", i, evaluated)
		}

		lenExpected := len(tt.expected)
		lenGot := len(result.Elements)

		if lenGot != lenExpected {
			t.Errorf("test %d: expected %d items, got=%d", i, lenExpected, lenGot)
		}

		for k, ei := range tt.expected {
			testIntegerObject(t, result.Elements[k], ei)
		}
	}
}

func TestArrayLiteralEmptyStrings(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{`[3]string{}`, []string{"", "", ""}},
		{`[5]string{"hello"}`, []string{"hello", "", "", "", ""}},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		result, ok := evaluated.(*object.Array)
		if !ok {
			t.Errorf("test %d: expected object.Array, got=%T", i, evaluated)
		}

		if len(result.Elements) != len(tt.expected) {
			t.Fatalf("test %d failed: expected %d elements, got=%d", i, len(tt.expected), len(result.Elements))
		}

		for k, ett := range tt.expected {
			str, ok := result.Elements[k].(*object.String)
			if !ok {
				t.Fatalf("test %d failed: expected object.String, got=%T", k, result.Elements[k])
			}

			if str.Value != ett {
				t.Fatalf("test %d failed: string %s does not match %s", k, str.Value, ett)
			}
		}
	}
}

func TestSliceLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected []int64
	}{
		{`[]int32{1, 2, 3}`, []int64{1, 2, 3}},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		result, ok := evaluated.(*object.Slice)
		if !ok {
			t.Errorf("test %d: expected object.Slice, got=%T", i, evaluated)
		}

		lenExpected := len(tt.expected)
		lenGot := len(result.Elements)

		if lenGot != lenExpected {
			t.Errorf("test %d: expected %d items, got=%d", i, lenExpected, lenGot)
		}

		for k, ei := range tt.expected {
			testIntegerObject(t, result.Elements[k], ei)
		}
	}
}

func TestMapLiterals(t *testing.T) {
	tests := []struct {
		input    string
		keys     []any
		values   []any
		numPairs int
	}{
		{`map[int32]string{ 10: "graham" }`, []any{10}, []any{"graham"}, 1},
		{`map[string]string{ "one": "one", "two": "two", "three":"three", }`, []any{"one", "two", "three"}, []any{"one", "two", "three"}, 3},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		m, ok := evaluated.(*object.Map)
		if !ok {
			t.Fatalf("test %d: expected object.Map, got=%T", i, evaluated)
		}

		if len(m.Pairs) != tt.numPairs {
			t.Fatalf("test %d: expected %d pairs in map, got %d", i, tt.numPairs, len(m.Pairs))
		}

		for k, key := range tt.keys {
			expectedKey := key
			expectedValue := tt.values[k]

			hashKey := hashKeyFromAny(t, expectedKey)

			value, ok := m.Pairs[hashKey]
			if !ok {
				t.Errorf("test %d: map does not contain key %v", i, expectedKey)
				continue
			}

			if value.Inspect() != fmt.Sprintf("%v", expectedValue) {
				t.Errorf("test %d: for key %v, expected value %v, got %s",
					i, expectedKey, expectedValue, value.Inspect())
			}
		}
	}
}

func TestAccessMapLiterals(t *testing.T) {
	tests := []struct {
		input  string
		target string
	}{
		{`map[int32]string{ 1: "one"}[1]`, "one"},
	}

	for i, tt := range tests {
		evaluated := testEval(t, tt.input)
		str, ok := evaluated.(*object.String)
		if !ok {
			t.Fatalf("test %d: expected object.String, got %T (%+v)", i, evaluated, evaluated)
		}

		if str.Value != tt.target {
			t.Errorf("test %d: expected string '%s' got %s", i, tt.target, str.Value)
		}
	}
}
