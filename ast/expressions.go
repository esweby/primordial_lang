package ast

import (
	"bytes"
	"strings"

	"github.com/esweby/primordial_lang/token"
)

// TupleTargetExpression is the parenthesized identifier list used on the
// left-hand side of tuple declarations and assignments.
type TupleTargetExpression struct {
	Token token.Token
	Names []*Identifier
}

func (tte *TupleTargetExpression) expressionNode() {}

func (tte *TupleTargetExpression) TokenLiteral() string {
	return tte.Token.Literal
}

func (tte *TupleTargetExpression) String() string {
	names := make([]string, 0, len(tte.Names))
	for _, name := range tte.Names {
		names = append(names, name.String())
	}
	return "(" + strings.Join(names, ", ") + ")"
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

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode() {}

func (sl *StringLiteral) TokenLiteral() string {
	return sl.Token.Literal
}

func (sl *StringLiteral) String() string {
	return sl.Token.Literal
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

type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode() {}

func (ce *CallExpression) TokenLiteral() string {
	return ce.Token.Literal
}

func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := make([]string, 0, len(ce.Arguments))
	for _, arg := range ce.Arguments {
		args = append(args, arg.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
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
