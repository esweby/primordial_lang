package ast

import (
	"bytes"

	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

type MapLiteral struct {
	Token token.Token
	Type  types.Type
	Pairs []*MapLiteralPair
}

func (ml *MapLiteral) expressionNode() {}

func (ml *MapLiteral) TokenLiteral() string {
	return ml.Token.Literal
}

func (ml *MapLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(ml.Type.Name())
	out.WriteString("{")

	for _, p := range ml.Pairs {
		out.WriteString("\n")
		out.WriteString(p.String())
	}

	out.WriteString("\n}")
	return out.String()
}

type MapLiteralPair struct {
	Token token.Token
	Key   Expression
	Value Expression
}

func (mlp *MapLiteralPair) TokenLiteral() string {
	return mlp.Token.Literal
}

func (mlp *MapLiteralPair) String() string {
	return mlp.Key.String() + ": " + mlp.Value.String() + ","
}
