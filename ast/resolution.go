package ast

import "github.com/esweby/primordial_lang/token"

// Symbol is the view of a resolved name that later phases see on the AST.
// It is implemented by the resolver package; the interface lives here to
// avoid an import cycle between ast and resolver.
type Symbol interface {
	Name() string
	// Declaration returns the node that introduced the name, e.g. a
	// *DeclareStatement, *FunctionStatement, *Parameter or *StructStatement.
	// Builtins have no declaration and return nil.
	Declaration() Node
}

// Scope is the view of a lexical scope that later phases see on the AST.
// It is implemented by the resolver package.
type Scope interface {
	Parent() Scope
	Lookup(name string) (Symbol, bool)
}

// TokenOf returns the token a node starts at, for error positions. Nodes
// without a token return the zero token.
func TokenOf(node Node) token.Token {
	switch n := node.(type) {
	case *IntegerLiteral:
		return n.Token
	case *Identifier:
		return n.Token
	case *PrefixExpression:
		return n.Token
	case *InfixExpression:
		return n.Token
	case *CallExpression:
		return n.Token
	case *IfExpression:
		return n.Token
	case *BlockExpression:
		return n.Token
	case *ForLoop:
		return n.Token
	case *IndexExpression:
		return n.Token
	case *MemberExpression:
		return n.Token
	case *StringLiteral:
		return n.Token
	case *Boolean:
		return n.Token
	case *ArrayLiteral:
		return n.Token
	case *SliceLiteral:
		return n.Token
	case *MapLiteral:
		return n.Token
	case *StructLiteral:
		return n.Token
	case *FunctionLiteral:
		return n.Token
	case *DeclareStatement:
		return n.Token
	case *TupleDeclareStatement:
		return n.Token
	case *AssignStatement:
		return n.Token
	case *TupleAssignStatement:
		return n.Token
	case *ReturnStatement:
		return n.Token
	case *ExpressionStatement:
		return n.Token
	case *FunctionStatement:
		return n.Token
	case *StructStatement:
		return n.Token
	case *BreakStatement:
		return n.Token
	case *ContinueStatement:
		return n.Token
	default:
		return token.Token{}
	}
}
