package ast

import (
	"bytes"
	"strings"

	"github.com/esweby/primordial_lang/token"
)

type ForLoop struct {
	Token token.Token
	Label *Identifier
	Controller ForController
	Body *BlockExpression
}

func (fl *ForLoop) expressionNode() {}

func (fl *ForLoop) TokenLiteral() string {
	return fl.Token.Literal
}

func (fl *ForLoop) String() string {
	var out bytes.Buffer

	if fl.Label != nil {
		out.WriteString(fl.Label.String())
		out.WriteString(": ")
	}

	out.WriteString("for ")
	out.WriteString(fl.Controller.String())
	out.WriteString(fl.Body.String())

	return out.String()
}

type ForController interface {
	Node
	controller()
}

type Infinite struct {
	Token token.Token
}

func (i *Infinite) controller() {}

func (i *Infinite) TokenLiteral() string { return i.Token.Literal }

func (i *Infinite) String() string { return "" }

type While struct {
	Token token.Token
	Condition Expression
}

func (w *While) controller() {}

func (w *While) TokenLiteral() string { return w.Token.Literal }

func (w *While) String() string {
	return "(" + w.Condition.String() + ") "
}

type Constructed struct {
	Token token.Token
	Initializer Statement
	Condition Expression
	Iterator Expression
}

func (c *Constructed) controller() {}

func (c *Constructed) TokenLiteral() string { return c.Token.Literal }

func (c *Constructed) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(c.Initializer.String())
	out.WriteString("; ")
	out.WriteString(c.Condition.String())
	out.WriteString("; ")
	out.WriteString(c.Iterator.String())
	out.WriteString(") ")

	return out.String()
}

type Range struct {
	Token token.Token
	Variables []*Identifier // length 1 or 2
    Iterable Expression
}

func (r *Range) controller() {}

func (r *Range) TokenLiteral() string { return r.Token.Literal }

func (r *Range) String() string {
	var out bytes.Buffer

	variables := []string{}
	for _, str := range r.Variables {
		variables = append(variables, str.String())
	}

	out.WriteString(strings.Join(variables, ", "))
	out.WriteString(" := range ")
	out.WriteString(r.Iterable.String())
	out.WriteString(" ")

	return out.String()
}
