package types

import (
	"bytes"
	"math/big"
	"strconv"
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
	KindSlice
	KindFunction
	KindTuple
	KindMap
	KindStruct
	KindUntypedInteger
)

type Type interface {
	Name() string
	Size() int
	Kind() Kind
}

type Indexable interface{
	Type
	IndexType() Type
	ElementType() Type
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

type Integer struct {
	typeName string
	bits     uint16
	signed   bool
}

func NewInteger(name string, bits uint16, signed bool) *Integer {
	return &Integer{typeName: name, bits: bits, signed: signed}
}

func (i *Integer) Name() string { return i.typeName }
func (i *Integer) Size() int    { return int(i.bits / 8) }
func (i *Integer) Kind() Kind   { return KindInteger }
func (i *Integer) Bits() uint16 { return i.bits }
func (i *Integer) Signed() bool { return i.signed }

func (i *Integer) MinValue() *big.Int {
	if !i.signed {
		return new(big.Int)
	}
	return new(big.Int).Neg(new(big.Int).Lsh(big.NewInt(1), uint(i.bits-1)))
}

func (i *Integer) MaxValue() *big.Int {
	bits := i.bits
	if i.signed {
		bits--
	}
	return new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), uint(bits)), big.NewInt(1))
}

func (i *Integer) CanRepresent(value *big.Int) bool {
	return value != nil && value.Cmp(i.MinValue()) >= 0 && value.Cmp(i.MaxValue()) <= 0
}

type UntypedInteger struct{}

func (u *UntypedInteger) Name() string { return "untyped integer" }
func (u *UntypedInteger) Size() int    { return 0 }
func (u *UntypedInteger) Kind() Kind   { return KindUntypedInteger }

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

type Function struct {
	ParamTypes  []Type
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
func (fn *Function) Size() int  { return 16 }
func (fn *Function) Kind() Kind { return KindFunction }
func NewFunction(paramTypes, returnTypes []Type) *Function {
	return &Function{ParamTypes: paramTypes, ReturnTypes: returnTypes}
}

type Array struct {
	elementType Type
	length      int
}

func (al *Array) ElementType() Type { return al.elementType }
func (al *Array) Length() int { return al.length }
func (al *Array) Size() int   { return al.length * al.elementType.Size() }
func (al *Array) Kind() Kind  { return KindArray }
func (al *Array) Name() string {
	return "[" + strconv.Itoa(al.length) + "]" + al.elementType.Name()
}
func NewArray(t Type, length int) *Array {
	return &Array{elementType: t, length: length}
}

type Slice struct {
	elementType Type
	length      int
}

func (sl *Slice) ElementType() Type { return sl.elementType }
func (sl *Slice) Length() int { return sl.length }
func (sl *Slice) Size() int   { return sl.length * sl.elementType.Size() }
func (sl *Slice) Kind() Kind  { return KindSlice }
func (sl *Slice) Name() string {
	return "[]" + sl.elementType.Name()
}
func (sl *Slice) IndexType() Type { return Int64Type }

func NewSlice(t Type) *Slice {
	return &Slice{elementType: t, length: 0}
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
func (n *Named) LookupMember(name string) (MemberDefinition, bool) {
	provider, ok := n.Underlying.(MemberProvider)
	if !ok {
		return MemberDefinition{}, false
	}
	return provider.LookupMember(name)
}

func (n *Named) LookupTypeMember(name string) (MemberDefinition, bool) {
	provider, ok := n.Underlying.(TypeMemberProvider)
	if !ok {
		return MemberDefinition{}, false
	}
	return provider.LookupTypeMember(name)
}

type Map struct {
	Key Type
	Value Type
}

func (m *Map) Name() string { return "map[" + m.Key.Name() + "]" + m.Value.Name() }
func (m *Map) Size() int    { return 16 }
func (m *Map) Kind() Kind   { return KindMap }
func (m *Map) IndexType() Type { return m.Key }
func (m *Map) ElementType() Type { return m.Value }

type StructField struct {
	Name       string
	Type       Type
	HasDefault bool
	Public     bool
}

type Struct struct {
	TypeName      string
	Fields        map[string]StructField
	TypeFunctions map[string]MemberDefinition
	Methods       map[string]MemberDefinition
}

func (s *Struct) Name() string { return s.TypeName }
func (s *Struct) Size() int    { return 0 }
func (s *Struct) Kind() Kind   { return KindStruct }

func (s *Struct) LookupMember(name string) (MemberDefinition, bool) {
	if field, ok := s.Fields[name]; ok {
		return MemberDefinition{
			Name:        field.Name,
			Kind:        MemberProperty,
			ReturnTypes: []Type{field.Type},
			Public:      field.Public,
			StructOwner: s,
		}, true
	}

	method, ok := s.Methods[name]
	return method, ok
}

func (s *Struct) LookupTypeMember(name string) (MemberDefinition, bool) {
	function, ok := s.TypeFunctions[name]
	return function, ok
}
