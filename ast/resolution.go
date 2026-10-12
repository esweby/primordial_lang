package ast

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
