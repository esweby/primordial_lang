package object

import (
	"bytes"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

type ObjectType string

type Type interface {
	Name() string
}

type BuiltinFunction func(arg ...Object) Object

const (
	INTEGER_OBJ           = "INTEGER"
	BOOLEAN_OBJ           = "BOOLEAN"
	STRING_OBJ            = "STRING"
	FUNCTION_OBJ          = "FUNCTION"
	RETURN_VALUES_OBJ     = "RETURN"
	TUPLE_OBJ             = "TUPLE"
	ERROR_OBJ             = "ERROR"
	BUILTIN_OBJ           = "BUILTIN"
	ARRAY_OBJ             = "ARRAY"
	SLICE_OBJ             = "SLICE"
	STRUCT_DEFINITION_OBJ = "STRUCT_DEFINITION"
	STRUCT_OBJ            = "STRUCT"
	MAP_OBJ				  = "MAP"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Hasheable interface {
	HashKey() HashKey
}

type HashKey struct {
	Type ObjectType
	Value string
}

type Integer struct {
	Value       *big.Int
	IntegerType types.Type
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string { return i.Value.String() }
func (i *Integer) HashKey() HashKey {
    return HashKey{Type: i.Type(), Value: i.Inspect()}
}

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) HashKey() HashKey {
    return HashKey{Type: b.Type(), Value: b.Inspect()}
}

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string { return s.Value }
func (s *String) HashKey() HashKey {
    return HashKey{Type: s.Type(), Value: s.Value}
}

type ReturnValue struct {
	Value []Object
}

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUES_OBJ }
func (rv *ReturnValue) Inspect() string {
	if len(rv.Value) == 1 {
		return rv.Value[0].Inspect()
	}

	values := []string{}
	for _, v := range rv.Value {
		values = append(values, v.Inspect())
	}

	return "(" + strings.Join(values, ", ") + ")"
}

type Error struct {
	Message string
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string { return "ERROR: " + e.Message }


type Function struct {
	Name        string
	Parameters  []*ast.Parameter
	ReturnTypes []*ast.ReturnType
	Body        *ast.BlockExpression
	Env         *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var out bytes.Buffer

	out.WriteString("fn")

	if f.Name != "" {
		out.WriteString(" ")
		out.WriteString(f.Name)
	}

	params := make([]string, 0, len(f.Parameters))
	for _, parameter := range f.Parameters {
		params = append(params, parameter.String())
	}

	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")

	if len(f.ReturnTypes) > 0 {
		returns := make([]string, 0, len(f.ReturnTypes))
		for _, returnType := range f.ReturnTypes {
			returns = append(returns, returnType.String())
		}

		out.WriteString(": ")
		out.WriteString(strings.Join(returns, ", "))
	}

	out.WriteString(" { ")
	out.WriteString(f.Body.String())
	out.WriteString(" }")

	return out.String()
}

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	elements := make([]string, 0, len(a.Elements))
	for _, e := range a.Elements {
		elements = append(elements, e.Inspect())
	}

	return "{" + strings.Join(elements, ", ") + "}"
}

type Slice struct {
	Elements []Object
}

func (s *Slice) Type() ObjectType { return SLICE_OBJ }
func (s *Slice) Inspect() string {
	elements := make([]string, 0, len(s.Elements))
	for _, e := range s.Elements {
		elements = append(elements, e.Inspect())
	}

	return "{" + strings.Join(elements, ", ") + "}"
}

type Tuple struct {
	Elements []Object
}

func (t *Tuple) Type() ObjectType { return TUPLE_OBJ }
func (t *Tuple) Inspect() string {
	elements := make([]string, 0, len(t.Elements))
	for _, e := range t.Elements {
		elements = append(elements, e.Inspect())
	}

	return "(" + strings.Join(elements, ", ") + ")"
}

type Builtin struct {
	Fn BuiltinFunction
}

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ }
func (b *Builtin) Inspect() string { return "builtin function" }

type StructDefinition struct {
	Declaration *ast.StructStatement
	Env         *Environment
}

func (sd *StructDefinition) Type() ObjectType { return STRUCT_DEFINITION_OBJ }
func (sd *StructDefinition) Inspect() string  { return sd.Declaration.String() }

type Struct struct {
	Name       string
	Definition *StructDefinition
	Fields     map[string]Object
}

func (s *Struct) Type() ObjectType { return STRUCT_OBJ }
func (s *Struct) Inspect() string {
	names := make([]string, 0, len(s.Fields))
	for name := range s.Fields {
		names = append(names, name)
	}
	sort.Strings(names)

	fields := make([]string, 0, len(names))
	for _, name := range names {
		fields = append(fields, name+": "+s.Fields[name].Inspect())
	}
	return s.Name + "{" + strings.Join(fields, ", ") + "}"
}

type Map struct {
	Pairs map[HashKey]Object
}

func (m *Map) Type() ObjectType { return MAP_OBJ }
func (m *Map) Inspect() string {
    keys := make([]HashKey, 0, len(m.Pairs))
    for key := range m.Pairs {
        keys = append(keys, key)
    }

    sort.Slice(keys, func(i, j int) bool {
        if keys[i].Type != keys[j].Type {
            return keys[i].Type < keys[j].Type
        }
        return keys[i].Value < keys[j].Value
    })

    fields := make([]string, 0, len(keys))
    for _, key := range keys {
        fields = append(fields, key.Value+": "+m.Pairs[key].Inspect())
    }

    return "{" + strings.Join(fields, ", ") + "}"
}