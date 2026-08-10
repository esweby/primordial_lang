package evaluator

import (
	"testing"

	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/types"
)

func TestEvaluatedIntegersPreserveTheirTypes(t *testing.T) {
	tests := []struct {
		input    string
		expected types.Type
	}{
		{`value: int8 := 1; value;`, types.Int8Type},
		{`value: int32 := 1; value + 2;`, types.Int32Type},
		{`value: uint64 := 18446744073709551615; value;`, types.UInt64Type},
		{`value := 1; value;`, types.Int64Type},
		{`1 + 2;`, types.Int64Type},
	}
	for _, test := range tests {
		integer, ok := testEval(test.input).(*object.Integer)
		if !ok {
			t.Fatalf("expected integer for %q", test.input)
		}
		if !types.IsTypesEqual(integer.IntegerType, test.expected) {
			t.Errorf("expected %s, got %s", test.expected.Name(), integer.IntegerType.Name())
		}
	}
}

func TestCheckedIntegerOverflow(t *testing.T) {
	tests := []string{
		`value: int8 := 127; value + 1;`,
		`value: int8 := -128; value - 1;`,
		`value: uint8 := 255; value + 1;`,
		`value: uint8 := 0; value - 1;`,
		`value: int8 := 64; value * 2;`,
		`value: int8 := -128; value / -1;`,
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			errorObject, ok := testEval(input).(*object.Error)
			if !ok {
				t.Fatalf("expected overflow error, got %T", testEval(input))
			}
			if errorObject.Message == "" {
				t.Fatal("expected overflow message")
			}
		})
	}
}

func TestCheckedRuntimeIntegerConversion(t *testing.T) {
	converted := testEval(`value: int64 := 255; uint8(value);`)
	integer, ok := converted.(*object.Integer)
	if !ok || !types.IsTypesEqual(integer.IntegerType, types.UInt8Type) {
		t.Fatalf("expected uint8 conversion, got %#v", converted)
	}

	overflow := testEval(`value: int64 := 256; uint8(value);`)
	if _, ok := overflow.(*object.Error); !ok {
		t.Fatalf("expected conversion overflow, got %#v", overflow)
	}
}

func TestArrayNeutralIntegersPreserveElementType(t *testing.T) {
	array, ok := testEval(`[2]uint16{};`).(*object.Array)
	if !ok {
		t.Fatal("expected array")
	}
	for _, element := range array.Elements {
		integer := element.(*object.Integer)
		if !types.IsTypesEqual(integer.IntegerType, types.UInt16Type) {
			t.Fatalf("expected uint16 neutral element, got %s", integer.IntegerType.Name())
		}
	}
}
