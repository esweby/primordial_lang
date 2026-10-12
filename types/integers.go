package types

import (
	"fmt"
	"math/big"
)

// IntType and UIntType are distinct 64-bit types, following Go. They are not
// interchangeable with int64 and uint64.
var IntType = NewInteger("int", 64, true)
var UIntType = NewInteger("uint", 64, false)

// DefaultIntegerType is the type an untyped integer constant takes when
// nothing in its statement gives it a type.
var DefaultIntegerType = IntType

// FoldIntegerConstant evaluates a binary operator on two integer constants
// with arbitrary precision. ok is false when the operator does not produce an
// integer constant.
func FoldIntegerConstant(operator string, left, right *big.Int) (*big.Int, bool, error) {
	switch operator {
	case "+":
		return new(big.Int).Add(left, right), true, nil
	case "-":
		return new(big.Int).Sub(left, right), true, nil
	case "*":
		return new(big.Int).Mul(left, right), true, nil
	case "/":
		if right.Sign() == 0 {
			return nil, true, fmt.Errorf("division by zero")
		}
		return new(big.Int).Quo(left, right), true, nil
	default:
		return nil, false, nil
	}
}

// CheckRepresentable reports an error if value does not fit in target.
func CheckRepresentable(value *big.Int, target Type) error {
	integer, ok := target.(*Integer)
	if !ok {
		return fmt.Errorf("cannot use untyped integer constant as %s", target.Name())
	}
	if !integer.CanRepresent(value) {
		return fmt.Errorf("integer constant %s overflows %s", value.String(), target.Name())
	}
	return nil
}
