package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (sa *SemanticAnalyzer) analyzeFunctionStatement(stmt *ast.FunctionStatement) {
	// Build the function signature from the AST.
	paramTypes := make([]types.Type, len(stmt.Parameters))
	for i, p := range stmt.Parameters {
		paramTypes[i] = p.Type
	}
	returnTypes := make([]types.Type, len(stmt.ReturnTypes))
	for i, rt := range stmt.ReturnTypes {
		returnTypes[i] = rt.Type
	}

	// Register the function in the current scope with its signature.
	sa.current.Set(stmt.Name.Value, &FunctionSymbol{
		name:        stmt.Name.Value,
		typ:         types.FunctionType, // generic, but signature stored separately
		params:      stmt.Parameters,
		returnTypes: returnTypes,
	})

	// Enter function scope.
	sa.enterScope()

	// Add parameters to the new scope.
	for _, p := range stmt.Parameters {
		sa.current.Set(p.Name.Value, &BasicSymbol{
			name: p.Name.Value,
			typ:  p.Type,
		})
	}

	previousReturnTypes := sa.returnTypes
	sa.returnTypes = returnTypes
	defer func() {
		sa.returnTypes = previousReturnTypes
	}()

	// Analyze the body.
	bodyResult := sa.analyzeBlock(stmt.Body)

	// Validate return paths.
	if len(returnTypes) > 0 {
		if !bodyResult.Returns {
			sa.error(fmt.Sprintf("function '%s' declares return types but does not return on all paths",
				stmt.Name.Value))
		}
	} else {
		// Void function: if it returns unexpectedly, we catch it in analyzeReturnStatement.
		// No further action.
	}

	// Exit function scope.
	sa.exitScope()
}

func (sa *SemanticAnalyzer) analyzeFunctionLiteral(ds *ast.DeclareStatement) types.Type {
	fnLit, ok := ds.Value.(*ast.FunctionLiteral)
	if !ok {
		sa.error("analyzeFunctionLiteral called with non-function value")
		return types.InvalidType
	}

	// Build parameter symbols and return types
	params := make([]*ast.Parameter, len(fnLit.Parameters))
	copy(params, fnLit.Parameters)

	returnTypes := make([]types.Type, len(fnLit.ReturnTypes))
	for i, rt := range fnLit.ReturnTypes {
		returnTypes[i] = rt.Type
	}

	// Register the function symbol in the current scope
	sa.current.Set(ds.Name.Value, &FunctionSymbol{
		name:        ds.Name.Value,
		typ:         types.FunctionType,
		params:      params,
		returnTypes: returnTypes,
	})

	// Enter function scope, add parameters
	sa.enterScope()
	for _, p := range fnLit.Parameters {
		sa.current.Set(p.Name.Value, &BasicSymbol{name: p.Name.Value, typ: p.Type})
	}

	previousReturnTypes := sa.returnTypes
	sa.returnTypes = returnTypes
	defer func() {
		sa.returnTypes = previousReturnTypes
	}()

	for _, stmt := range fnLit.Body.Statements {
		sa.analyzeStatement(stmt)
	}

	sa.exitScope()
	return types.FunctionType
}

func (sa *SemanticAnalyzer) analyzeStandaloneFunctionLiteral(fnLit *ast.FunctionLiteral) {
	// Build signature for context.
	returnTypes := make([]types.Type, len(fnLit.ReturnTypes))
	for i, rt := range fnLit.ReturnTypes {
		returnTypes[i] = rt.Type
	}

	sa.enterScope()
	for _, p := range fnLit.Parameters {
		sa.current.Set(p.Name.Value, &BasicSymbol{
			name: p.Name.Value,
			typ:  p.Type,
		})
	}

	previousReturnTypes := sa.returnTypes
	sa.returnTypes = returnTypes
	defer func() {
		sa.returnTypes = previousReturnTypes
	}()

	bodyResult := sa.analyzeBlock(fnLit.Body)
	if len(returnTypes) > 0 && !bodyResult.Returns {
		sa.error("function literal declares return types but does not return on all paths")
	}

	sa.exitScope()
}
