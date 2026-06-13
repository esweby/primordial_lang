package types

import (
	"testing"
)

func TestGetBuiltin(t *testing.T) {
	tests := []struct {
		input    string
		wantType Type
		wantOk   bool
	}{
		{"bool", &Bool{}, true},
		{"int8", &Int8{}, true},
		{"uint8", &UInt8{}, true},
		{"int16", &Int16{}, true},
		{"uint16", &UInt16{}, true},
		{"int32", &Int32{}, true},
		{"uint32", &UInt32{}, true},
		{"int64", &Int64{}, true},
		{"uint64", &UInt64{}, true},
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
