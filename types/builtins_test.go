package types

import (
	"math/big"
	"testing"
)

func TestGetBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		wantType Type
		wantOk   bool
	}{
		{"bool", &Bool{}, true},
		{"int8", Int8Type, true},
		{"uint8", UInt8Type, true},
		{"int16", Int16Type, true},
		{"uint16", UInt16Type, true},
		{"int32", Int32Type, true},
		{"uint32", UInt32Type, true},
		{"int64", Int64Type, true},
		{"uint64", UInt64Type, true},
		{"float32", &Float32{}, true},
		{"float64", &Float64{}, true},
		{"string", &String{}, true},
		{"unknown", nil, false},
	}

	for _, tt := range tests {
		got, ok := GetBuiltin(tt.input)
		if ok != tt.wantOk {
			t.Errorf("GetBuiltin(%q) ok = %v, want %v", tt.input, ok, tt.wantOk)
		}
		if !typesEqual(got, tt.wantType) {
			t.Errorf("GetBuiltin(%q) = %T, want %T", tt.input, got, tt.wantType)
		}
	}
}

func TestNeutralValue(t *testing.T) {
	tests := []struct {
		name string
		typ  Type
		want any
		ok   bool
	}{
		{"integer", Int32Type, big.NewInt(0), true},
		{"unsigned integer", UInt64Type, big.NewInt(0), true},
		{"float", Float32Type, float64(0), true},
		{"string", StringType, "", true},
		{"boolean", BoolType, false, true},
		{"array", NewArray(Int32Type, 1), nil, false},
		{"nil", nil, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := NeutralValue(tt.typ)
			valuesEqual := got == tt.want
			if gotInteger, ok := got.(*big.Int); ok {
				wantInteger, wantOK := tt.want.(*big.Int)
				valuesEqual = wantOK && gotInteger.Cmp(wantInteger) == 0
			}
			if ok != tt.ok || !valuesEqual {
				t.Fatalf("NeutralValue(%v) = (%#v, %t), want (%#v, %t)", tt.typ, got, ok, tt.want, tt.ok)
			}
		})
	}
}

// typesEqual checks if two Types are the same concrete type
// (since we don't have equality for interfaces).
func typesEqual(a, b Type) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// Compare by name (unique for built‑ins)
	return a.Name() == b.Name()
}

func TestIntegerMetadataAndBounds(t *testing.T) {
	tests := []struct {
		typ     *Integer
		bits    uint16
		signed  bool
		minimum string
		maximum string
	}{
		{Int8Type, 8, true, "-128", "127"},
		{UInt8Type, 8, false, "0", "255"},
		{Int16Type, 16, true, "-32768", "32767"},
		{UInt16Type, 16, false, "0", "65535"},
		{Int32Type, 32, true, "-2147483648", "2147483647"},
		{UInt32Type, 32, false, "0", "4294967295"},
		{Int64Type, 64, true, "-9223372036854775808", "9223372036854775807"},
		{UInt64Type, 64, false, "0", "18446744073709551615"},
	}
	for _, test := range tests {
		t.Run(test.typ.Name(), func(t *testing.T) {
			if test.typ.Bits() != test.bits || test.typ.Signed() != test.signed {
				t.Fatalf("unexpected metadata for %s", test.typ.Name())
			}
			if test.typ.MinValue().String() != test.minimum || test.typ.MaxValue().String() != test.maximum {
				t.Fatalf("unexpected bounds for %s: %s..%s", test.typ.Name(), test.typ.MinValue(), test.typ.MaxValue())
			}
			minimum, _ := new(big.Int).SetString(test.minimum, 10)
			maximum, _ := new(big.Int).SetString(test.maximum, 10)
			if !test.typ.CanRepresent(minimum) || !test.typ.CanRepresent(maximum) {
				t.Fatalf("%s should represent both boundaries", test.typ.Name())
			}
		})
	}
}
