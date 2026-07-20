package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

type ArrayLiteral struct {
	Token token.Token
	Type types.Type
	Size int
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode() {}

func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }

func (al *ArrayLiteral) String() string { 
	var out bytes.Buffer

	elements := []string{}
	for _, el := range al.Elements {
		elements = append(elements, el.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
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