package types

type Kind int

const (
	KindInvalid Kind = iota
	KindInteger
	KindFloat
	KindString
	KindBoolean
	KindArray
	KindFunction
)

type Type interface {
	Name() string
	Size() int
	Kind() Kind
}

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

type Named struct {
	CustomName string
	Underlying Type
}

func (n *Named) Name() string { return n.CustomName }
func (n *Named) Size() int    { return n.Underlying.Size() }
func (n *Named) Kind() Kind   { return n.Underlying.Kind() }
