package types

import "math/big"

var InvalidType = &Invalid{}
var BoolType = &Bool{}
var Int8Type = NewInteger("int8", 8, true)
var Int16Type = NewInteger("int16", 16, true)
var Int32Type = NewInteger("int32", 32, true)
var Int64Type = NewInteger("int64", 64, true)
var UInt8Type = NewInteger("uint8", 8, false)
var UInt16Type = NewInteger("uint16", 16, false)
var UInt32Type = NewInteger("uint32", 32, false)
var UInt64Type = NewInteger("uint64", 64, false)
var UntypedIntegerType = &UntypedInteger{}
var Float32Type = &Float32{}
var Float64Type = &Float64{}
var StringType = &String{}
var FunctionType = &Function{}
var TupleType = &Tuple{}
var VoidType = &Void{}

var builtins = map[string]Type{
	"invalid":  InvalidType,
	"bool":     BoolType,
	"int8":     Int8Type,
	"uint8":    UInt8Type,
	"int16":    Int16Type,
	"uint16":   UInt16Type,
	"int32":    Int32Type,
	"uint32":   UInt32Type,
	"int64":    Int64Type,
	"uint64":   UInt64Type,
	"float32":  Float32Type,
	"float64":  Float64Type,
	"string":   StringType,
	"function": FunctionType,
	"void":     VoidType,
}

func GetBuiltin(typeName string) (Type, bool) {
	typ, ok := builtins[typeName]
	return typ, ok
}

func NeutralValue(t Type) (any, bool) {
	if t == nil {
		return nil, false
	}

	switch t.Kind() {
	case KindInteger:
		return big.NewInt(0), true
	case KindFloat:
		return float64(0), true
	case KindString:
		return "", true
	case KindBoolean:
		return false, true
	default:
		return nil, false
	}
}

// Maybe a better name
func IsTypesEqual(a, b Type) bool {
	if a == nil || b == nil {
		return false
	}

	return a.Name() == b.Name()
}

func IsInvalid(t Type) bool {
	if t == nil {
		return true
	}

	return t.Kind() == KindInvalid
}

func IsInteger(t Type) bool {
	return t != nil && t.Kind() == KindInteger
}

func IsUntypedInteger(t Type) bool {
	return t != nil && t.Kind() == KindUntypedInteger
}

func IsFloat(t Type) bool {
	return t.Kind() == KindFloat
}

func IsNumeric(t Type) bool {
	return t != nil && (t.Kind() == KindInteger || t.Kind() == KindUntypedInteger || t.Kind() == KindFloat)
}

func IsString(t Type) bool {
	return t.Kind() == KindString
}

func IsBoolean(t Type) bool {
	return t.Kind() == KindBoolean
}

func IsArrayLiteral(t Type) bool {
	return t.Kind() == KindArray
}

func IsFunction(t Type) bool {
	return t.Kind() == KindFunction
}

func GetFunctionSignature(t Type) ([]Type, []Type, bool) {
	if f, ok := t.(*Function); ok {
		return f.ParamTypes, f.ReturnTypes, true
	}

	return nil, nil, false
}

func IsTuple(t Type) bool {
	return t.Kind() == KindTuple
}

func IsMap(t Type) bool {
	return t.Kind() == KindMap
}

func IsVoid(t Type) bool {
	return t.Kind() == KindVoid
}

func IsAssignable(to, from Type) bool {
	if to == nil || from == nil {
		return false
	}
	return IsTypesEqual(to, from)
}

func UnwrapTuple(t Type) ([]Type, bool) {
	if tup, ok := t.(*Tuple); ok {
		return tup.Types, true
	}

	return nil, false
}
