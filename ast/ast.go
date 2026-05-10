package ast

import (
	"bytes"
	"strings"

	"github.com/esweby/primordial_lang/token"
)

type Node interface {
	TokenLiteral() string
	String() string
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

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode() {}

func (i *Identifier) TokenLiteral() string {
	return i.Token.Literal
}

func (i *Identifier) String() string { return i.Value }

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
	out.WriteString(" := ")

	if dl.Value != nil {
		out.WriteString(dl.Value.String())
	}

	out.WriteString(";")

	return out.String()
}

type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode() {}

func (rs *ReturnStatement) TokenLiteral() string {
	return rs.Token.Literal
}

func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
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

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}

func (il *IntegerLiteral) TokenLiteral() string {
	return il.Token.Literal
}

func (il *IntegerLiteral) String() string {
	return il.Token.Literal
}

type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode() {}

func (pe *PrefixExpression) TokenLiteral() string {
	return pe.Token.Literal
}

func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode() {}

func (ie *InfixExpression) TokenLiteral() string {
	return ie.Token.Literal
}

func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" ")
	out.WriteString(ie.Operator)
	out.WriteString(" ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode() {}

func (b *Boolean) TokenLiteral() string {
	return b.Token.Literal
}

func (b *Boolean) String() string {
	return b.Token.Literal
}

// Issue with this implementation is that it is ignorant of if. if else, else
// Todo would be to swap Token token.Token for a pos argument and
type IfExpression struct {
	Token     token.Token
	Condition Expression
	Body      *BlockExpression
	Else      Expression
}

func (ife *IfExpression) expressionNode() {}

func (ife *IfExpression) TokenLiteral() string {
	return ife.Token.Literal
}

func (ife *IfExpression) String() string {
	var out bytes.Buffer

	if ife.Condition != nil {
		out.WriteString("if ")
	} else {
		out.WriteString("else ")
	}
	out.WriteString(ife.Condition.String())
	out.WriteString(" ")
	out.WriteString(ife.Body.String())
	if ife.Else != nil {
		out.WriteString(ife.Else.String())
	}

	return out.String()
}

type FunctionLiteral struct {
	Token      token.Token
	Parameters []*Parameter
	ReturnTypes []*ReturnType
	Body       *BlockExpression
}

func (fl *FunctionLiteral) expressionNode() {}

func (fl *FunctionLiteral) TokenLiteral() string {
	return fl.Token.Literal
}

func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.String())
	}

	returnTypes := []string{}
	for _, rt := range fl.ReturnTypes {
		returnTypes = append(returnTypes, rt.String())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString("): ")
	out.WriteString(strings.Join(returnTypes, ", "))
	out.WriteString(fl.Body.String())

	return out.String()
}

type Parameter struct {
	Name *Identifier
	Type string
}

func (p *Parameter) String() string {
	var out bytes.Buffer
	out.WriteString(p.Name.String())
	out.WriteString(" ")
	out.WriteString(p.Type)
	return out.String()
}

type ReturnType struct {
	Type string
}

func (rt *ReturnType) String() string {
	return rt.Type
}

type BlockExpression struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockExpression) expressionNode() {}

func (bs *BlockExpression) TokenLiteral() string {
	return bs.Token.Literal
}

func (bs *BlockExpression) String() string {
	var out bytes.Buffer

	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}
