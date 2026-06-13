package types

var InvalidType = &Invalid{}
var BoolType = &Bool{}
var Int8Type = &Int8{}
var Int16Type = &Int16{}
var Int32Type = &Int32{}
var Int64Type = &Int64{}
var UInt8Type = &UInt8{}
var UInt16Type = &UInt16{}
var UInt32Type = &UInt32{}
var UInt64Type = &UInt64{}
var Float32Type = &Float32{}
var Float64Type = &Float64{}
var StringType = &String{}
var FunctionType = &Function{}

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
}

func GetBuiltin(typeName string) (Type, bool) {
	typ, ok := builtins[typeName]
	return typ, ok
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
	return t.Kind() == KindInteger
}

func IsFloat(t Type) bool {
	return t.Kind() == KindFloat
}

func IsNumeric(t Type) bool {
	return t.Kind() == KindInteger || t.Kind() == KindFloat
}

func IsString(t Type) bool {
	return t.Kind() == KindString
}

func IsBoolean(t Type) bool {
	return t.Kind() == KindBoolean
}

func IsArray(t Type) bool {
	return t.Kind() == KindArray
}

func IsFunction(t Type) bool {
	return t.Kind() == KindFunction
}
