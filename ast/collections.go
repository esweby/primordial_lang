package ast

import (
	"bytes"

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


	return out.String()
}