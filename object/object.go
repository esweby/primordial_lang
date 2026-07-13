package object

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/esweby/primordial_lang/ast"
)

type ObjectType string

type Type interface {
	Name() string
}

type BuiltinFunction func(arg ...Object) Object

const (
	INTEGER_OBJ       = "INTEGER"
	BOOLEAN_OBJ       = "BOOLEAN"
	STRING_OBJ 		  = "STRING"
	FUNCTION_OBJ      = "FUNCTION"
	RETURN_VALUES_OBJ = "RETURN"
	TUPLE_OBJ         = "TUPLE"
	ERROR_OBJ         = "ERROR"
	BUILTIN_OBJ		  = "BUILTIN"
)

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

type Boolean struct {
	Value bool
}

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

type String struct {
	Value string
}

func (s *String) Type() ObjectType { return STRING_OBJ }

func (s *String) Inspect() string { return s.Value }

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

func (b *Builtin) Type() ObjectType { return BUILTIN_OBJ; }

func (b *Builtin) Inspect() string { return "builtin function"}
