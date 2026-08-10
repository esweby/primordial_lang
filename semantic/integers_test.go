package semantic

import (
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/types"
)

func TestIntegerLiteralBoundaries(t *testing.T) {
	valid := []string{
		`value: int8 := -128;`,
		`value: int8 := 127;`,
		`value: uint8 := 0;`,
		`value: uint8 := 255;`,
		`value: int64 := -9223372036854775808;`,
		`value: int64 := 9223372036854775807;`,
		`value: uint64 := 18446744073709551615;`,
	}
	for _, input := range valid {
		t.Run(input, func(t *testing.T) {
			_, errors := analyzeProgram(t, input)
			if len(errors) != 0 {
				t.Fatalf("expected valid integer boundary, got %v", errors)
			}
		})
	}

	invalid := []string{
		`value: int8 := -129;`,
		`value: int8 := 128;`,
		`value: uint8 := -1;`,
		`value: uint8 := 256;`,
		`value: int64 := -9223372036854775809;`,
		`value: int64 := 9223372036854775808;`,
		`value: uint64 := 18446744073709551616;`,
	}
	for _, input := range invalid {
		t.Run(input, func(t *testing.T) {
			_, errors := analyzeProgram(t, input)
			if !containsSemanticError(errors, "not representable") {
				t.Fatalf("expected representability error, got %v", errors)
			}
		})
	}
}

func TestIntegerLiteralUsesContextSymmetrically(t *testing.T) {
	analyzer, errors := analyzeProgram(t, `
		x: int32 := 1;
		left := x + 2;
		right := 2 + x;
		comparison := x < 3;
	`)
	if len(errors) != 0 {
		t.Fatalf("expected contextual integer resolution, got %v", errors)
	}
	for _, name := range []string{"x", "left", "right"} {
		symbol, ok := analyzer.Symbols().Get(name)
		if !ok || !types.IsTypesEqual(symbol.Type(), types.Int32Type) {
			t.Fatalf("expected %s to be int32, got %#v", name, symbol)
		}
	}
}

func TestContextFreeIntegerDefaultsToInt64(t *testing.T) {
	analyzer, errors := analyzeProgram(t, `value := 1 + 2;`)
	if len(errors) != 0 {
		t.Fatalf("expected default integer materialization, got %v", errors)
	}
	symbol, _ := analyzer.Symbols().Get("value")
	if !types.IsTypesEqual(symbol.Type(), types.Int64Type) {
		t.Fatalf("expected int64 default, got %s", symbol.Type().Name())
	}
}

func TestRejectMixedConcreteIntegerTypes(t *testing.T) {
	_, errors := analyzeProgram(t, `x: int32 := 1; y: int64 := 2; result := x + y;`)
	if !containsSemanticError(errors, "mismatched integer types: int32 and int64") {
		t.Fatalf("expected concrete integer mismatch, got %v", errors)
	}
}

func TestRejectUnsignedNegation(t *testing.T) {
	_, errors := analyzeProgram(t, `x: uint32 := 1; result := -x;`)
	if !containsSemanticError(errors, "cannot negate unsigned integer uint32") {
		t.Fatalf("expected unsigned negation error, got %v", errors)
	}
}

func TestIntegerConversions(t *testing.T) {
	analyzer, errors := analyzeProgram(t, `
		x: int32 := 1;
		y := int64(x);
		z := uint8(255);
	`)
	if len(errors) != 0 {
		t.Fatalf("expected explicit conversions to analyze, got %v", errors)
	}
	for name, expected := range map[string]types.Type{"y": types.Int64Type, "z": types.UInt8Type} {
		symbol, _ := analyzer.Symbols().Get(name)
		if !types.IsTypesEqual(symbol.Type(), expected) {
			t.Errorf("expected %s to be %s, got %s", name, expected.Name(), symbol.Type().Name())
		}
	}

	_, errors = analyzeProgram(t, `value := uint8(-1);`)
	if !containsSemanticError(errors, "not representable as uint8") {
		t.Fatalf("expected conversion range error, got %v", errors)
	}
	_, errors = analyzeProgram(t, `value := int32(true);`)
	if len(errors) == 0 || !strings.Contains(errors[0].Error(), "cannot convert bool to int32") {
		t.Fatalf("expected non-integer conversion error, got %v", errors)
	}
}

func TestIntegerContextsAcrossLanguageFeatures(t *testing.T) {
	_, errors := analyzeProgram(t, `
		struct Packet { code: uint8 = 255; }
		fn identity(value int16): int16 { return value; }
		values := []int8{-128, 127};
		packet := Packet{};
		result := identity(12);
	`)
	if len(errors) != 0 {
		t.Fatalf("expected integer literals to materialize from all contexts, got %v", errors)
	}
}
