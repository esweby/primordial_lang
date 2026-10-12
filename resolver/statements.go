package resolver

import (
	"fmt"
	"slices"

	"github.com/esweby/primordial_lang/ast"
)

func (r *Resolver) resolveStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		r.resolveExpression(s.Expression)
	case *ast.DeclareStatement:
		r.resolveDeclareStatement(s)
	case *ast.TupleDeclareStatement:
		r.resolveTupleDeclareStatement(s)
	case *ast.AssignStatement:
		r.resolveAssignStatement(s)
	case *ast.TupleAssignStatement:
		r.resolveTupleAssignStatement(s)
	case *ast.ReturnStatement:
		for _, value := range s.ReturnValues {
			r.resolveExpression(value)
		}
	case *ast.FunctionStatement:
		r.resolveFunctionStatement(s)
	case *ast.StructStatement:
		r.resolveStructStatement(s)
	case *ast.BreakStatement:
		r.resolveBreakStatement(s)
	case *ast.ContinueStatement:
		r.resolveContinueStatement(s)
	case nil:
		return
	default:
		r.errors = append(r.errors, fmt.Sprintf("resolver received unexpected statement: %T", stmt))
	}
}

func (r *Resolver) resolveDeclareStatement(stmt *ast.DeclareStatement) {
	r.resolveType(stmt.Type, stmt.Token)

	// A function literal can refer to itself, so its name is declared before
	// the body is resolved. Any other value is resolved first, so that
	// `x := x + 1` refers to an outer x.
	if fnLit, ok := stmt.Value.(*ast.FunctionLiteral); ok {
		sym := r.declare(stmt.Name, VariableSymbol, stmt)
		body := r.resolveFunctionLiteral(fnLit)
		if sym != nil {
			sym.Body = body
		}
		return
	}

	r.resolveExpression(stmt.Value)
	r.declare(stmt.Name, VariableSymbol, stmt)
}

func (r *Resolver) resolveTupleDeclareStatement(stmt *ast.TupleDeclareStatement) {
	r.resolveExpression(stmt.Value)

	seen := make(map[string]struct{}, len(stmt.Names))
	for _, name := range stmt.Names {
		if name.Value == "_" {
			continue
		}
		if _, duplicate := seen[name.Value]; duplicate {
			r.errorAt(name.Token, "variable '%s' appears more than once in tuple declaration", name.Value)
			continue
		}
		seen[name.Value] = struct{}{}
		r.declare(name, TupleElementSymbol, stmt)
	}
}

func (r *Resolver) resolveAssignStatement(stmt *ast.AssignStatement) {
	if stmt.Name != nil {
		r.lookup(stmt.Name)
	} else {
		r.resolveExpression(stmt.Target)
	}
	r.resolveExpression(stmt.Value)
}

func (r *Resolver) resolveTupleAssignStatement(stmt *ast.TupleAssignStatement) {
	for _, name := range stmt.Names {
		r.lookup(name)
	}
	r.resolveExpression(stmt.Value)
}

// resolveBlock resolves a block in a new scope owned by the block.
func (r *Resolver) resolveBlock(block *ast.BlockExpression) {
	if block == nil {
		return
	}
	block.Scope = r.enterScope(block)
	defer r.exitScope()
	r.resolveStatements(block.Statements)
}

func (r *Resolver) resolveStatements(stmts []ast.Statement) {
	for _, stmt := range stmts {
		r.resolveStatement(stmt)
	}
}

// resolveFunctionStatement declares the function before resolving its body so
// that it can call itself.
func (r *Resolver) resolveFunctionStatement(stmt *ast.FunctionStatement) {
	sym := r.declare(stmt.Name, FunctionSymbol, stmt)
	body := r.resolveFunction(stmt, stmt.Token, stmt.Parameters, stmt.ReturnTypes, stmt.Body, nil)
	stmt.Scope = body
	if sym != nil {
		sym.Body = body
	}
}

// predeclareStructs declares every struct in a statement list up front so
// that structs can be referenced before their declaration.
func (r *Resolver) predeclareStructs(stmts []ast.Statement) {
	for _, stmt := range stmts {
		if s, ok := stmt.(*ast.StructStatement); ok {
			r.predeclared[s] = struct{}{}
			r.declare(s.Name, StructSymbol, s)
		}
	}
}

func (r *Resolver) resolveStructStatement(stmt *ast.StructStatement) {
	// Structs outside the top level are not predeclared.
	if _, predeclared := r.predeclared[stmt]; !predeclared {
		r.declare(stmt.Name, StructSymbol, stmt)
	}

	// Field names are members, not names in scope. Types and default values
	// are resolved in the scope enclosing the struct.
	for _, field := range stmt.Fields {
		r.resolveType(field.Type, field.Token)
		r.resolveExpression(field.Value)
	}

	for _, function := range stmt.TypeFunctions {
		function.Scope = r.resolveFunction(function, function.Token, function.Parameters, function.ReturnTypes, function.Body, nil)
	}

	if stmt.Impl != nil {
		for _, method := range stmt.Impl.Methods {
			method.Scope = r.resolveFunction(method, method.Token, method.Parameters, method.ReturnTypes, method.Body, stmt)
		}
	}
}

func (r *Resolver) resolveBreakStatement(stmt *ast.BreakStatement) {
	if r.loopDepth == 0 {
		r.errorAt(stmt.Token, "break outside of loop")
	} else if stmt.Label != nil && !slices.Contains(r.loopLabels, stmt.Label.Value) {
		r.errorAt(stmt.Label.Token, "break label '%s' does not match any enclosing loop", stmt.Label.Value)
	}

	r.resolveExpression(stmt.Value)
}

func (r *Resolver) resolveContinueStatement(stmt *ast.ContinueStatement) {
	if r.loopDepth == 0 {
		r.errorAt(stmt.Token, "continue outside of loop")
	}
}
