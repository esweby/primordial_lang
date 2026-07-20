package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (sa *SemanticAnalyzer) analyzeTupleDeclareStatement(stmt *ast.TupleDeclareStatement) types.Type {
	rhsType := sa.analyzeExpression(stmt.Value)
	elementTypes, ok := types.UnwrapTuple(rhsType)
	if !ok {
		sa.error(fmt.Sprintf("tuple declaration requires a tuple value, got %s", rhsType.Name()))
		return types.InvalidType
	}

	if len(stmt.Names) != len(elementTypes) {
		sa.error(fmt.Sprintf("tuple declaration arity mismatch: expected %d names, got %d",
			len(elementTypes), len(stmt.Names)))
		return types.InvalidType
	}

	seen := make(map[string]struct{})
	for _, name := range stmt.Names {
		if name.Value == "_" {
			continue
		}
		if _, duplicate := seen[name.Value]; duplicate {
			sa.error(fmt.Sprintf("variable '%s' appears more than once in tuple declaration", name.Value))
			return types.InvalidType
		}
		seen[name.Value] = struct{}{}

		if sa.current.ExistsInCurrentScope(name.Value) {
			sa.error(fmt.Sprintf("variable '%s' already declared in this scope", name.Value))
			return types.InvalidType
		}
	}

	for i, name := range stmt.Names {
		if name.Value == "_" {
			continue
		}
		sa.current.Set(name.Value, &DeclareSymbol{
			name: name.Value,
			typ:  elementTypes[i],
			mut:  false,
		})
	}

	return rhsType
}

func (sa *SemanticAnalyzer) analyzeTupleAssignmentStatement(stmt *ast.TupleAssignStatement) types.Type {
	rhsType := sa.analyzeExpression(stmt.Value)
	elementTypes, ok := types.UnwrapTuple(rhsType)
	if !ok {
		sa.error(fmt.Sprintf("tuple assignment requires a tuple value, got %s", rhsType.Name()))
		return types.InvalidType
	}

	if len(stmt.Names) != len(elementTypes) {
		sa.error(fmt.Sprintf("tuple assignment arity mismatch: expected %d names, got %d",
			len(elementTypes), len(stmt.Names)))
		return types.InvalidType
	}

	for i, name := range stmt.Names {
		if name.Value == "_" {
			continue
		}

		sym, found := sa.current.Get(name.Value)
		if !found {
			sa.error(fmt.Sprintf("undefined variable: %s", name.Value))
			return types.InvalidType
		}

		decl, isVariable := sym.(*DeclareSymbol)
		if !isVariable {
			sa.error(fmt.Sprintf("cannot assign to non-variable: %s", name.Value))
			return types.InvalidType
		}
		if !decl.Mutable() {
			sa.error(fmt.Sprintf("cannot assign to immutable variable: %s", name.Value))
			return types.InvalidType
		}
		if !types.IsAssignable(decl.Type(), elementTypes[i]) {
			sa.error(fmt.Sprintf("tuple assignment value %d: expected %s, got %s",
				i, decl.Type().Name(), elementTypes[i].Name()))
			return types.InvalidType
		}
	}

	return rhsType
}
