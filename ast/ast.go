package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/esweby/primordial_lang/token"
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

	returnTypes := []string{}
	for _, rt := range fl.ReturnTypes {
		returnTypes = append(returnTypes, rt.String())
	}

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

	args := []string{}
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

func parametersToString(params []*Parameter) string {
	var out bytes.Buffer

	paramsString := []string{}
	for _, p := range params {
		paramsString = append(paramsString, p.String())
	}

	out.WriteString("(")
	out.WriteString(strings.Join(paramsString, ", "))
	out.WriteString(")")

	return out.String()
}

func returnTypesToString(returnTyoes []*ReturnType) string {
	var out bytes.Buffer

	returnTypeStrings := []string{}
	for _, rt := range returnTyoes {
		returnTypeStrings = append(returnTypeStrings, rt.String())
	}

	if len(returnTypeStrings) > 0 {
		out.WriteString(": ")
		out.WriteString(strings.Join(returnTypeStrings, ", "))
	}
	out.WriteString(" ")

	return out.String()
}
