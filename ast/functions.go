package ast

import (
	"bytes"
	"strings"

	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

type FunctionStatement struct {
	Token       token.Token
	Public      bool
	Name        *Identifier
	Parameters  []*Parameter
	ReturnTypes []*ReturnType
	Body        *BlockExpression
}

func (fs *FunctionStatement) statementNode() {}

func (fs *FunctionStatement) TokenLiteral() string {
	return fs.Token.Literal
}

func (fs *FunctionStatement) String() string {
	var out bytes.Buffer

	if fs.Public {
		out.WriteString("pub ")
	}

	out.WriteString("fn ")
	out.WriteString(fs.Name.String())
	out.WriteString(parametersToString(fs.Parameters))
	out.WriteString(returnTypesToString(fs.ReturnTypes))
	out.WriteString(fs.Body.String())

	return out.String()
}

type FunctionLiteral struct {
	Token       token.Token
	Parameters  []*Parameter
	ReturnTypes []*ReturnType
	Body        *BlockExpression
}

func (fl *FunctionLiteral) expressionNode() {}

func (fl *FunctionLiteral) TokenLiteral() string {
	return fl.Token.Literal
}

func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	out.WriteString(fl.TokenLiteral())
	out.WriteString(parametersToString(fl.Parameters))
	out.WriteString(returnTypesToString(fl.ReturnTypes))
	out.WriteString(fl.Body.String())

	return out.String()
}

type Parameter struct {
	Token token.Token
	Name  *Identifier
	Type  types.Type
}

func (p *Parameter) expressionNode() {}

func (p *Parameter) TokenLiteral() string {
	return p.Token.Literal
}

func (p *Parameter) String() string {
	var out bytes.Buffer
	out.WriteString(p.Name.String())
	out.WriteString(" ")
	out.WriteString(p.Type.Name())
	return out.String()
}

type ReturnType struct {
	Type types.Type
}

func (rt *ReturnType) String() string {
	return rt.Type.Name()
}

func parametersToString(params []*Parameter) string {
	paramsString := make([]string, 0, len(params))
	for _, p := range params {
		paramsString = append(paramsString, p.String())
	}

	return "(" + strings.Join(paramsString, ", ") + ")"
}

func returnTypesToString(returnTypes []*ReturnType) string {
	returnTypeStrings := make([]string, 0, len(returnTypes))
	for _, rt := range returnTypes {
		returnTypeStrings = append(returnTypeStrings, rt.String())
	}

	if len(returnTypeStrings) == 0 {
		return " "
	}

	return ": " + strings.Join(returnTypeStrings, ", ") + " "
}
