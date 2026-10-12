package ast

import (
	"bytes"

	"github.com/esweby/primordial_lang/types"
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

type TypedExpression interface {
	Expression
	GetResolvedType() types.Type
	SetResolvedType(types.Type)
}

// Typed holds the type the inference phase assigns to an expression. It is
// embedded in every expression node.
type Typed struct {
	ResolvedType types.Type
}

func (t *Typed) GetResolvedType() types.Type    { return t.ResolvedType }
func (t *Typed) SetResolvedType(typ types.Type) { t.ResolvedType = typ }

type Assignable interface {
	Expression
	isAssignable()
}

type Program struct {
	Statements []Statement
	Scope      Scope // set by the resolver
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
