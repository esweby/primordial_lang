package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (sa *SemanticAnalyzer) analyzeDeclareStatement(stmt *ast.DeclareStatement) types.Type {
	// Check redefinition
	if sa.current.ExistsInCurrentScope(stmt.Name.Value) {
		sa.error(fmt.Sprintf("variable '%s' already declared in this scope", stmt.Name.Value))
		return types.InvalidType
	}

	// Handle function literal assignment
	if _, ok := stmt.Value.(*ast.FunctionLiteral); ok {
		fnType := sa.analyzeFunctionLiteral(stmt) // analyzes body and registers symbol

		// If there is an explicit type annotation, it must be a function type.
		if stmt.Type != nil {
			if !types.IsFunction(stmt.Type) {
				sa.error(fmt.Sprintf("declaration type mismatch: expected %s, got function",
					stmt.Type.Name()))
				return types.InvalidType
			}
			// If annotation is `function`, we accept it.
		} else {
			// No annotation: infer as function type.
			stmt.SetInferredType(fnType)
		}
		return fnType
	}

	// Normal (non‑function) value analysis (existing code)
	rhsType := sa.analyzeExpression(stmt.Value)
	if types.IsInvalid(rhsType) {
		return types.InvalidType
	}

	if stmt.Type != nil && !types.IsAssignable(stmt.Type, rhsType) {
		sa.error(fmt.Sprintf("declaration type mismatch: expected %s, got %s",
			stmt.Type.Name(), rhsType.Name()))
		return types.InvalidType
	}

	if stmt.Type == nil {
		stmt.SetInferredType(rhsType)
	}

	sa.current.Set(stmt.Name.Value, &DeclareSymbol{
		name: stmt.Name.Value,
		typ:  stmt.GetType(),
		mut:  stmt.Mutable,
	})

	return stmt.GetType()
}

func (sa *SemanticAnalyzer) analyzeReturnStatement(stmt *ast.ReturnStatement) {
	// If we're not inside a function with return types, it's an error.
	if len(sa.returnTypes) == 0 {
		if len(stmt.ReturnValues) > 0 {
			sa.error("unexpected return with values in void function")
		}
		return
	}

	// Non-void function: must return the correct number and types.
	if len(stmt.ReturnValues) != len(sa.returnTypes) {
		sa.error(fmt.Sprintf("wrong number of return values: expected %d, got %d",
			len(sa.returnTypes), len(stmt.ReturnValues)))
		return
	}

	for i, rv := range stmt.ReturnValues {
		valType := sa.analyzeExpression(rv)
		expected := sa.returnTypes[i]
		if !types.IsAssignable(expected, valType) {
			sa.error(fmt.Sprintf("return value %d: expected %s, got %s",
				i, expected.Name(), valType.Name()))
		}
	}
}

func (sa *SemanticAnalyzer) analyzeAssignmentStatement(stmt *ast.AssignStatement) types.Type {
	// The LHS must be an identifier (for now).
	ident := stmt.Name
	if ident == nil {
		// handle error
		return types.InvalidType
	}

	// Look up the variable.
	sym, ok := sa.current.Get(ident.Value)
	if !ok {
		sa.error(fmt.Sprintf("undefined variable: %s", ident.Value))
		return types.InvalidType
	}

	// Check mutability.
	declSym, ok := sym.(*DeclareSymbol)
	if !ok {
		sa.error(fmt.Sprintf("cannot assign to non-variable: %s", ident.Value))
		return types.InvalidType
	}
	if !declSym.Mutable() {
		sa.error(fmt.Sprintf("cannot assign to immutable variable: %s", ident.Value))
		return types.InvalidType
	}

	// Analyze RHS.
	rhsType := sa.analyzeExpression(stmt.Value)
	if types.IsInvalid(rhsType) {
		return types.InvalidType
	}

	// Type check.
	if !types.IsTypesEqual(declSym.Type(), rhsType) {
		sa.error(fmt.Sprintf("assignment type mismatch: expected %s, got %s",
			declSym.Type().Name(), rhsType.Name()))
		return types.InvalidType
	}

	return rhsType
}
