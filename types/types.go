package types

import (
	"bytes"
	"strings"
)

type Kind int

const (
	KindInvalid Kind = iota
	KindVoid
	KindInteger
	KindFloat
	KindString
	KindBoolean
	KindArray
	KindFunction
	KindTuple
)

type Type interface {
	Name() string
	Size() int
	Kind() Kind
}

type Invalid struct{}

func (inv *Invalid) Name() string { return "invalid" }
func (inv *Invalid) Size() int    { return 0 }
func (inv *Invalid) Kind() Kind   { return KindInvalid }

type Void struct{}

func (v *Void) Name() string { return "void" }
func (v *Void) Size() int    { return 0 }
func (v *Void) Kind() Kind   { return KindVoid }

type Bool struct{}

func (b *Bool) Name() string { return "bool" }
func (b *Bool) Size() int    { return 1 }
func (b *Bool) Kind() Kind   { return KindBoolean }

type Int8 struct{}

func (i8 *Int8) Name() string { return "int8" }
func (i8 *Int8) Size() int    { return 1 }
func (i8 *Int8) Kind() Kind   { return KindInteger }

type Int16 struct{}

func (i16 *Int16) Name() string { return "int16" }
func (i16 *Int16) Size() int    { return 2 }
func (i16 *Int16) Kind() Kind   { return KindInteger }

type Int32 struct{}

func (i32 *Int32) Name() string { return "int32" }
func (i32 *Int32) Size() int    { return 4 }
func (i32 *Int32) Kind() Kind   { return KindInteger }

type Int64 struct{}

func (i64 *Int64) Name() string { return "int64" }
func (i64 *Int64) Size() int    { return 8 }
func (i64 *Int64) Kind() Kind   { return KindInteger }

type UInt8 struct{}

func (ui8 *UInt8) Name() string { return "uint8" }
func (ui8 *UInt8) Size() int    { return 1 }
func (ui8 *UInt8) Kind() Kind   { return KindInteger }

type UInt16 struct{}

func (ui16 *UInt16) Name() string { return "uint16" }
func (ui16 *UInt16) Size() int    { return 2 }
func (ui16 *UInt16) Kind() Kind   { return KindInteger }

type UInt32 struct{}

func (ui32 *UInt32) Name() string { return "uint32" }
func (ui32 *UInt32) Size() int    { return 4 }
func (ui32 *UInt32) Kind() Kind   { return KindInteger }

type UInt64 struct{}

func (ui64 *UInt64) Name() string { return "uint64" }
func (ui64 *UInt64) Size() int    { return 8 }
func (ui64 *UInt64) Kind() Kind   { return KindInteger }

type Float32 struct{}

func (fl32 *Float32) Name() string { return "float32" }
func (fl32 *Float32) Size() int    { return 4 }
func (fl32 *Float32) Kind() Kind   { return KindFloat }

type Float64 struct{}

func (fl64 *Float64) Name() string { return "float64" }
func (fl64 *Float64) Size() int    { return 8 }
func (fl64 *Float64) Kind() Kind   { return KindFloat }

type String struct{}

func (s *String) Name() string { return "string" }
func (s *String) Size() int    { return 16 }
func (s *String) Kind() Kind   { return KindString }

type Function struct{
	ParamTypes []Type
	ReturnTypes []Type
}

func (fn *Function) Name() string { 
	var out bytes.Buffer

	out.WriteString("fn(")
	params := []string{}
	for _, p := range fn.ParamTypes {
		params = append(params, p.Name())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString("): ")

	if len(fn.ReturnTypes) > 0 {
		rt := []string{}
		for _, r := range fn.ReturnTypes {
			rt = append(rt, r.Name())
		}
		out.WriteString(strings.Join(rt, ", "))
	} else {
		out.WriteString("void")
	}


	return out.String()
}
func (fn *Function) Size() int    { return 16 }
func (fn *Function) Kind() Kind   { return KindFunction }
func  NewFunction(paramTypes, returnTypes []Type) *Function {
	return &Function{ ParamTypes: paramTypes, ReturnTypes: returnTypes }
}

type Tuple struct {
	Types []Type
}

func (t *Tuple) Name() string {
	names := make([]string, len(t.Types))
	for i, typ := range t.Types {
		names[i] = typ.Name()
	}

	return "(" + strings.Join(names, ", ") + ")"
}

func (t *Tuple) Size() int {
	return 0
}

func (t *Tuple) Kind() Kind {
	return KindTuple
}

type Named struct {
	CustomName string
	Underlying Type
}

func (n *Named) Name() string { return n.CustomName }
func (n *Named) Size() int    { return n.Underlying.Size() }
func (n *Named) Kind() Kind   { return n.Underlying.Kind() }
