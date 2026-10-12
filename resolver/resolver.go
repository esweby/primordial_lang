package resolver

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

// builtinNames are resolvable without a declaration: len and the integer
// conversion functions, e.g. int32(x).
var builtinNames = []string{
	"len",
	"int", "int8", "int16", "int32", "int64",
	"uint", "uint8", "uint16", "uint32", "uint64",
}

// Resolver binds every identifier in a program to its declaration and records
// the lexical scopes on the AST. It does not look at types beyond checking
// that named types refer to declared structs.
type Resolver struct {
	program    *ast.Program
	universe   *Scope
	current    *Scope
	loopDepth  int
	loopLabels []string
	errors     []string

	predeclared map[*ast.StructStatement]struct{}
}

func New(program *ast.Program) *Resolver {
	universe := NewScope(nil)
	for _, name := range builtinNames {
		universe.Set(&Symbol{name: name, Kind: BuiltinSymbol})
	}

	return &Resolver{
		program:    program,
		universe:   universe,
		current:    universe,
		loopLabels: []string{},
		errors:     []string{},

		predeclared: make(map[*ast.StructStatement]struct{}),
	}
}

// Resolve walks the program and returns its top-level scope. The scope's
// parent is the universe scope holding the builtins.
func (r *Resolver) Resolve() (*Scope, []string) {
	if r.program == nil {
		r.errors = append(r.errors, "resolver started with nil program")
		return nil, r.errors
	}

	scope := r.enterScope(r.program)
	defer r.exitScope()
	r.program.Scope = scope

	r.predeclareStructs(r.program.Statements)
	for _, stmt := range r.program.Statements {
		r.resolveStatement(stmt)
	}

	return scope, r.errors
}

func (r *Resolver) Errors() []string { return r.errors }

func (r *Resolver) enterScope(node ast.Node) *Scope {
	r.current = NewChildScope(r.current, node)
	return r.current
}

func (r *Resolver) exitScope() {
	if r.current.parent != nil {
		r.current = r.current.parent
	}
}

// declare adds a symbol to the current scope, reporting a redeclaration if the
// name already exists in that same scope. "_" is never declared.
func (r *Resolver) declare(ident *ast.Identifier, kind SymbolKind, decl ast.Node) *Symbol {
	if ident == nil || ident.Value == "_" {
		return nil
	}

	if existing, exists := r.current.GetLocal(ident.Value); exists {
		r.errorAt(ident.Token, "%s '%s' already declared in this scope", existing.Kind, ident.Value)
		return nil
	}

	sym := &Symbol{name: ident.Value, Kind: kind, Ident: ident, Decl: decl}
	r.current.Set(sym)
	ident.Symbol = sym
	return sym
}

// lookup binds an identifier to the symbol it refers to.
func (r *Resolver) lookup(ident *ast.Identifier) (*Symbol, bool) {
	if ident == nil || ident.Value == "_" {
		return nil, false
	}

	sym, ok := r.current.Get(ident.Value)
	if !ok {
		r.errorAt(ident.Token, "undefined identifier: %s", ident.Value)
		return nil, false
	}

	ident.Symbol = sym
	return sym, true
}

func (r *Resolver) errorAt(tok token.Token, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if tok.Line > 0 {
		msg = fmt.Sprintf("%d:%d: %s", tok.Line, tok.Column, msg)
	}
	r.errors = append(r.errors, msg)
}

// resolveFunction opens one scope for the parameters and body together, so a
// body declaration cannot shadow a parameter. Loop context is reset because
// break and continue cannot cross a function boundary. When self is non-nil it
// is bound as the method receiver.
func (r *Resolver) resolveFunction(
	node ast.Node,
	tok token.Token,
	params []*ast.Parameter,
	returnTypes []*ast.ReturnType,
	body *ast.BlockExpression,
	self *ast.StructStatement,
) *Scope {
	for _, param := range params {
		r.resolveType(param.Type, param.Token)
	}
	for _, rt := range returnTypes {
		r.resolveType(rt.Type, tok)
	}

	scope := r.enterScope(node)
	defer r.exitScope()

	previousDepth, previousLabels := r.loopDepth, r.loopLabels
	r.loopDepth, r.loopLabels = 0, []string{}
	defer func() {
		r.loopDepth, r.loopLabels = previousDepth, previousLabels
	}()

	if self != nil {
		r.current.Set(&Symbol{name: "self", Kind: SelfSymbol, Decl: self})
	}
	for _, param := range params {
		r.declare(param.Name, ParameterSymbol, param)
	}

	if body != nil {
		body.Scope = scope
		r.resolveStatements(body.Statements)
	}

	return scope
}

// resolveType checks that every named type inside a type annotation refers to
// a declared struct, and links the named type to that struct's symbol.
func (r *Resolver) resolveType(typ types.Type, tok token.Token) {
	switch t := typ.(type) {
	case nil:
		return
	case *types.Named:
		sym, ok := r.current.Get(t.CustomName)
		if !ok {
			r.errorAt(tok, "unknown type: %s", t.CustomName)
			return
		}
		if !sym.IsType() {
			r.errorAt(tok, "%s is not a type", t.CustomName)
			return
		}
		t.Symbol = sym
	case *types.Array:
		r.resolveType(t.ElementType(), tok)
	case *types.Slice:
		r.resolveType(t.ElementType(), tok)
	case *types.Map:
		r.resolveType(t.Key, tok)
		r.resolveType(t.Value, tok)
	case *types.Tuple:
		for _, element := range t.Types {
			r.resolveType(element, tok)
		}
	case *types.Function:
		for _, param := range t.ParamTypes {
			r.resolveType(param, tok)
		}
		for _, ret := range t.ReturnTypes {
			r.resolveType(ret, tok)
		}
	}
}

// Scope is a lexical scope. It can be walked upwards through Parent and
// downwards through Children, and keeps symbols both in declaration order and
// indexed by name.
type Scope struct {
	parent   *Scope
	Children []*Scope
	// Node is the AST node that owns the scope (nil for the universe scope).
	Node    ast.Node
	Symbols []*Symbol
	byName  map[string]*Symbol
}

func NewScope(node ast.Node) *Scope {
	return &Scope{
		Node:     node,
		Children: []*Scope{},
		Symbols:  []*Symbol{},
		byName:   make(map[string]*Symbol),
	}
}

func NewChildScope(parent *Scope, node ast.Node) *Scope {
	sc := NewScope(node)
	sc.parent = parent
	if parent != nil {
		parent.Children = append(parent.Children, sc)
	}
	return sc
}

// Parent satisfies ast.Scope. It returns an untyped nil at the root so that
// callers can compare against nil.
func (s *Scope) Parent() ast.Scope {
	if s.parent == nil {
		return nil
	}
	return s.parent
}

func (s *Scope) ParentScope() *Scope { return s.parent }

// Lookup satisfies ast.Scope.
func (s *Scope) Lookup(name string) (ast.Symbol, bool) {
	sym, ok := s.Get(name)
	if !ok {
		return nil, false
	}
	return sym, true
}

// Get finds a symbol in this scope or any enclosing scope.
func (s *Scope) Get(name string) (*Symbol, bool) {
	for sc := s; sc != nil; sc = sc.parent {
		if sym, ok := sc.byName[name]; ok {
			return sym, true
		}
	}
	return nil, false
}

// GetLocal finds a symbol in this scope only.
func (s *Scope) GetLocal(name string) (*Symbol, bool) {
	sym, ok := s.byName[name]
	return sym, ok
}

// Set declares a symbol in this scope, replacing any existing symbol of the
// same name. Redeclaration checks are the resolver's responsibility.
func (s *Scope) Set(sym *Symbol) {
	if _, exists := s.byName[sym.name]; !exists {
		s.Symbols = append(s.Symbols, sym)
	} else {
		for i, existing := range s.Symbols {
			if existing.name == sym.name {
				s.Symbols[i] = sym
			}
		}
	}
	sym.Scope = s
	s.byName[sym.name] = sym
}

// FindAllSymbols searches this scope and every descendant scope for symbols
// with the given name, in declaration order.
func (s *Scope) FindAllSymbols(name string) ([]*Symbol, bool) {
	symbols := []*Symbol{}
	if sym, ok := s.byName[name]; ok {
		symbols = append(symbols, sym)
	}

	for _, child := range s.Children {
		found, _ := child.FindAllSymbols(name)
		symbols = append(symbols, found...)
	}

	return symbols, len(symbols) > 0
}

type SymbolKind int

const (
	VariableSymbol     SymbolKind = iota // declared by *ast.DeclareStatement
	TupleElementSymbol                   // declared by *ast.TupleDeclareStatement
	ParameterSymbol                      // declared by *ast.Parameter
	RangeSymbol                          // declared by *ast.Range
	FunctionSymbol                       // declared by *ast.FunctionStatement
	StructSymbol                         // declared by *ast.StructStatement
	SelfSymbol                           // receiver of a struct method; declared by *ast.StructStatement
	BuiltinSymbol                        // builtin function or conversion, no declaration
)

var symbolKindNames = [...]string{
	VariableSymbol:     "variable",
	TupleElementSymbol: "tuple element",
	ParameterSymbol:    "parameter",
	RangeSymbol:        "range variable",
	FunctionSymbol:     "function",
	StructSymbol:       "struct",
	SelfSymbol:         "self",
	BuiltinSymbol:      "builtin",
}

func (k SymbolKind) String() string { return symbolKindNames[k] }

// Symbol links a name to the node that declared it. Types are deliberately
// not stored here; they live on the declaration node and are filled in by
// later phases.
type Symbol struct {
	name string
	Kind SymbolKind
	// Ident is the identifier that introduced the name (nil for self and builtins).
	Ident *ast.Identifier
	// Decl is the declaring node (nil for builtins).
	Decl ast.Node
	// Scope is the scope the symbol was declared in.
	Scope *Scope
	// Body is the scope the symbol owns, if any (functions and function literals).
	Body *Scope
}

func (s *Symbol) Name() string { return s.name }

func (s *Symbol) Declaration() ast.Node { return s.Decl }

// IsType reports whether the symbol names a type rather than a value.
func (s *Symbol) IsType() bool { return s.Kind == StructSymbol }
