package ast

import (
	"bytes"

	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

type StructStatement struct {
	Token         token.Token
	Name          *Identifier
	Fields        []*StructField
	TypeFunctions []*FunctionStatement
	Impl          *StructImplBlock
}

func (ss *StructStatement) statementNode() {}

func (ss *StructStatement) TokenLiteral() string {
	return ss.Token.Literal
}

func (ss *StructStatement) String() string {
	var out bytes.Buffer

	out.WriteString("struct ")
	out.WriteString(ss.Name.String())
	out.WriteString(" {\n")

	for _, sf := range ss.Fields {
		out.WriteString(sf.String())
		out.WriteString("\n")
	}

	for _, sm := range ss.TypeFunctions {
		out.WriteString(sm.String())
		out.WriteString("\n")
	}

	if ss.Impl != nil && len(ss.Impl.Methods) > 0 {
		out.WriteString(ss.Impl.String())
	}

	out.WriteString("\n}")

	return out.String()
}

type StructField struct {
	Token    token.Token
	Name     *Identifier
	Public   bool
	Type     types.Type
	Value    Expression
	Inferred bool
}

func (sf *StructField) TokenLiteral() string {
	return sf.Token.Literal
}

func (sf *StructField) String() string {
	var out bytes.Buffer

	if sf.Public {
		out.WriteString("pub ")
	}
	out.WriteString(sf.Name.Value)
	if sf.Type != nil && !sf.Inferred {
		out.WriteString(": ")
		out.WriteString(sf.Type.Name())
	}

	if sf.Value != nil {
		out.WriteString(" = ")
		out.WriteString(sf.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

func (sf *StructField) SetInferredType(t types.Type) {
	sf.Type = t
	sf.Inferred = true
}

func (sf *StructField) GetType() types.Type {
	return sf.Type
}

type StructImplBlock struct {
	Token   token.Token
	Methods []*FunctionStatement
}

func (sib *StructImplBlock) statementNode() {}

func (sib *StructImplBlock) TokenLiteral() string {
	return sib.Token.Literal
}

func (sib *StructImplBlock) String() string {
	var out bytes.Buffer
	out.WriteString("impl {")
	for _, ibm := range sib.Methods {
		out.WriteString(ibm.String())
		out.WriteString("\n")
	}
	out.WriteString("}")

	return out.String()
}

type StructLiteral struct {
	Token  token.Token
	Name   *Identifier
	Fields []*StructLiteralField
}

func (sl *StructLiteral) expressionNode() {}

func (sl *StructLiteral) TokenLiteral() string {
	return sl.Token.Literal
}

func (sl *StructLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(sl.Name.String())
	out.WriteString(" {")

	if len(sl.Fields) > 0 {
		out.WriteString("\n")
		for _, field := range sl.Fields {
			out.WriteString(field.String())
			out.WriteString("\n")
		}
	}
	out.WriteString("}")

	return out.String()
}

type StructLiteralField struct {
	Token     token.Token
	Name      *Identifier
	Value     Expression
	Shorthand bool
}

func (slf *StructLiteralField) TokenLiteral() string {
	return slf.Token.Literal
}

func (slf *StructLiteralField) String() string {
	var out bytes.Buffer
	out.WriteString(slf.Name.Value)

	if !slf.Shorthand {
		out.WriteString(": ")
		out.WriteString(slf.Value.String())
	}

	out.WriteString(",")
	return out.String()
}
