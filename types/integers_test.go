package types

import (
	"math/big"
	"strings"
	"testing"
)

func TestIntIsDistinctFromInt64(t *testing.T) {
	if IsTypesEqual(IntType, Int64Type) || IsTypesEqual(UIntType, UInt64Type) {
		t.Fatal("int and uint must be distinct from int64 and uint64")
	}
	if IntType.Bits() != 64 || UIntType.Bits() != 64 || !IntType.Signed() || UIntType.Signed() {
		t.Fatal("int and uint must be 64-bit signed and unsigned")
	}
	if DefaultIntegerType != IntType {
		t.Fatal("untyped integer constants must default to int")
	}
}

func TestFoldIntegerConstant(t *testing.T) {
	tests := []struct {
		operator string
		left     int64
		right    int64
		want     int64
		ok       bool
		err      string
	}{
		{"+", 2, 3, 5, true, ""},
		{"-", 2, 3, -1, true, ""},
		{"*", 4, 3, 12, true, ""},
		{"/", 7, 2, 3, true, ""},
		{"/", 7, 0, 0, true, "division by zero"},
		{"<", 1, 2, 0, false, ""},
	}

	for _, tt := range tests {
		got, ok, err := FoldIntegerConstant(tt.operator, big.NewInt(tt.left), big.NewInt(tt.right))
		if ok != tt.ok {
			t.Errorf("%d %s %d: ok = %v, want %v", tt.left, tt.operator, tt.right, ok, tt.ok)
			continue
		}
		if tt.err != "" {
			if err == nil || !strings.Contains(err.Error(), tt.err) {
				t.Errorf("%d %s %d: expected error %q, got %v", tt.left, tt.operator, tt.right, tt.err, err)
			}
			continue
		}
		if ok && got.Int64() != tt.want {
			t.Errorf("%d %s %d = %s, want %d", tt.left, tt.operator, tt.right, got, tt.want)
		}
	}
}

func TestCheckRepresentable(t *testing.T) {
	if err := CheckRepresentable(big.NewInt(127), Int8Type); err != nil {
		t.Errorf("127 should fit int8: %v", err)
	}
	if err := CheckRepresentable(big.NewInt(300), Int8Type); err == nil || !strings.Contains(err.Error(), "overflows int8") {
		t.Errorf("300 should overflow int8, got %v", err)
	}
	if err := CheckRepresentable(big.NewInt(-1), UIntType); err == nil {
		t.Error("-1 should not fit uint")
	}
	if err := CheckRepresentable(big.NewInt(1), StringType); err == nil || !strings.Contains(err.Error(), "as string") {
		t.Errorf("an integer constant should not convert to string, got %v", err)
	}
}
