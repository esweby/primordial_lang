package ast

import "github.com/esweby/primordial_lang/token"

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}

	return ""
}

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode() {}

func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

type DeclareStatement struct {
	Token    token.Token
	Name     *Identifier
	Value    Expression
	Mutable  bool
	Public   bool
	Constant bool
	Type     string
}

func (dl *DeclareStatement) statementNode() {}

func (dl *DeclareStatement) TokenLiteral() string {
	return dl.Token.Literal
}
