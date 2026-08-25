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

	if stmt.Type != nil {
		resolved := sa.resolveType(stmt.Type)
		if !types.IsInvalid(resolved) {
			stmt.Type = resolved
		}
	}

	// Normal (non‑function) value analysis (existing code)
	rhsType := sa.analyzeExpression(stmt.Value)
	if types.IsInvalid(rhsType) {
		return types.InvalidType
	}

	if stmt.Type != nil {
		if err := sa.requireAssignable(stmt.Type, rhsType, stmt.Value); err != nil {
			sa.error("declaration type mismatch: " + err.Error())
			return types.InvalidType
		}
	} else {
		resolvedType, err := sa.defaultInteger(stmt.Value, rhsType)
		if err != nil {
			sa.error(err.Error())
			return types.InvalidType
		}
		rhsType = resolvedType
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
		if err := sa.requireAssignable(expected, valType, rv); err != nil {
			sa.error(fmt.Sprintf("return value %d: %s", i, err.Error()))
		}
	}
}

func (sa *SemanticAnalyzer) analyzeAssignmentStatement(stmt *ast.AssignStatement) types.Type {
	ident := stmt.Name
	if ident != nil {
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
		if err := sa.requireAssignable(declSym.Type(), rhsType, stmt.Value); err != nil {
			sa.error("assignment type mismatch: " + err.Error())
			return types.InvalidType
		}
	
		return rhsType
	}

	switch target := stmt.Target.(type) {
	case *ast.MemberExpression:
		member, ok := sa.resolveMember(target)
		if !ok {
			return types.InvalidType
		}
	
		if member.Kind != types.MemberProperty || len(member.ReturnTypes) != 1 {
			sa.error(fmt.Sprintf("cannot assign to method: %s", member.Name))
			return types.InvalidType
		}

		rhsType := sa.analyzeExpression(stmt.Value)
		if types.IsInvalid(rhsType) {
			return types.InvalidType
		}
		if err := sa.requireAssignable(member.ReturnTypes[0], rhsType, stmt.Value); err != nil {
			sa.error("assignment type mismatch: " + err.Error())
			return types.InvalidType
		}
		return member.ReturnTypes[0]
	case *ast.IndexExpression:
		elementType := sa.analyzeIndexExpression(target)
		if types.IsInvalid(elementType) {
			return types.InvalidType
		}

		rhsType := sa.analyzeExpression(stmt.Value)
		if types.IsInvalid(rhsType) {
			return types.InvalidType
		}

		if err := sa.requireAssignable(elementType, rhsType, stmt.Value); err != nil {
			sa.error("assignment type mismatch: " + err.Error())
			return types.InvalidType
		}

		return elementType
	default:
		sa.error("invalid target assignment")
		return types.InvalidType
	}
}
