package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

type DeclareStatement struct {
	Token    token.Token
	Name     *Identifier
	Value    Expression
	Mutable  bool
	Public   bool
	Constant bool
	Type     types.Type
	Inferred bool
}

func (dl *DeclareStatement) statementNode() {}

func (dl *DeclareStatement) TokenLiteral() string {
	return dl.Token.Literal
}

func (dl *DeclareStatement) String() string {
	var out bytes.Buffer

	if dl.Public {
		out.WriteString("pub ")
	}

	if dl.Mutable {
		out.WriteString("mut ")
	} else if dl.Constant {
		out.WriteString("const ")
	}

	out.WriteString(dl.Name.String())

	if dl.Type != nil && !dl.Inferred {
		out.WriteString(": ")
		out.WriteString(dl.Type.Name())
	}
	out.WriteString(" := ")

	if dl.Value != nil {
		out.WriteString(dl.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

func (dl *DeclareStatement) SetInferredType(t types.Type) {
	dl.Inferred = true
	dl.Type = t
}

func (dl *DeclareStatement) GetType() types.Type {
	return dl.Type
}

type AssignStatement struct {
	Token token.Token
	Name  *Identifier
	Value Expression
}

func (as *AssignStatement) statementNode() {}

func (as *AssignStatement) TokenLiteral() string {
	return as.Token.Literal
}

func (as *AssignStatement) String() string {
	return fmt.Sprintf("%s = %s", as.Name.String(), as.Value.String())
}

// TupleDeclareStatement destructures a tuple value into newly declared names.
// An identifier named "_" discards the corresponding tuple element.
type TupleDeclareStatement struct {
	Token token.Token
	Names []*Identifier
	Value Expression
}

func (tds *TupleDeclareStatement) statementNode() {}

func (tds *TupleDeclareStatement) TokenLiteral() string {
	return tds.Token.Literal
}

func (tds *TupleDeclareStatement) String() string {
	names := make([]string, 0, len(tds.Names))
	for _, name := range tds.Names {
		names = append(names, name.String())
	}

	return fmt.Sprintf("(%s) := %s;", strings.Join(names, ", "), tds.Value.String())
}

// TupleAssignStatement destructures a tuple value into existing mutable names.
// An identifier named "_" discards the corresponding tuple element.
type TupleAssignStatement struct {
	Token token.Token
	Names []*Identifier
	Value Expression
}

func (tas *TupleAssignStatement) statementNode() {}

func (tas *TupleAssignStatement) TokenLiteral() string {
	return tas.Token.Literal
}

func (tas *TupleAssignStatement) String() string {
	names := make([]string, 0, len(tas.Names))
	for _, name := range tas.Names {
		names = append(names, name.String())
	}

	return fmt.Sprintf("(%s) = %s;", strings.Join(names, ", "), tas.Value.String())
}

type ReturnStatement struct {
	Token        token.Token
	ReturnValues []Expression
}

func (rs *ReturnStatement) statementNode() {}

func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	for i, rv := range rs.ReturnValues {
		if rv != nil {
			out.WriteString(rv.String())
			if i < len(rs.ReturnValues)-1 {
				out.WriteString(", ")
			}
		}
	}

	out.WriteString(";")

	return out.String()
}

type ExpressionStatement struct {
	Token      token.Token // first token
	Expression Expression
}

func (es *ExpressionStatement) statementNode() {}

func (es *ExpressionStatement) TokenLiteral() string {
	return es.Token.Literal
}

func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}
