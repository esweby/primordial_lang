package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (sa *SemanticAnalyzer) analyzeStructStatement(stmt *ast.StructStatement) types.Type {
	startErrors := len(sa.errors)
	structType, predeclared := sa.structDeclarations[stmt]
	if !predeclared {
		structType = sa.predeclareStruct(stmt)
	}
	if structType == nil {
		return types.InvalidType
	}

	for _, field := range stmt.Fields {
		if _, duplicate := structType.Fields[field.Name.Value]; duplicate {
			sa.error(fmt.Sprintf("field '%s' already declared on %s", field.Name.Value, stmt.Name.Value))
			continue
		}

		var defaultType types.Type
		if field.Value != nil {
			defaultType = sa.analyzeExpression(field.Value)
		}

		fieldType := sa.resolveType(field.Type)
		if field.Type == nil {
			if field.Value == nil {
				sa.error(fmt.Sprintf("field %s.%s requires a type or default value", stmt.Name.Value, field.Name.Value))
				continue
			}
			if types.IsInvalid(defaultType) {
				continue
			}
			resolvedType, err := sa.defaultInteger(field.Value, defaultType)
			if err != nil {
				sa.error(fmt.Sprintf("default value for %s.%s: %s", stmt.Name.Value, field.Name.Value, err.Error()))
				continue
			}
			fieldType = resolvedType
			field.SetInferredType(fieldType)
		} else if types.IsInvalid(fieldType) {
			sa.error(fmt.Sprintf("unknown type %s for field %s.%s", field.Type.Name(), stmt.Name.Value, field.Name.Value))
			continue
		}
		if !field.Inferred {
			field.Type = fieldType
		}

		if field.Value != nil {
			if err := sa.requireAssignable(fieldType, defaultType, field.Value); err != nil {
				sa.error(fmt.Sprintf("default value for %s.%s: %s", stmt.Name.Value, field.Name.Value, err.Error()))
			}
		}

		structType.Fields[field.Name.Value] = types.StructField{
			Name:       field.Name.Value,
			Type:       fieldType,
			HasDefault: field.Value != nil,
			Public:     field.Public,
		}
	}

	for _, function := range stmt.TypeFunctions {
		if _, duplicate := structType.TypeFunctions[function.Name.Value]; duplicate {
			sa.error(fmt.Sprintf("type function '%s' already declared on %s", function.Name.Value, stmt.Name.Value))
			continue
		}
		structType.TypeFunctions[function.Name.Value] = sa.structFunctionSignature(function)
	}

	if stmt.Impl != nil {
		for _, method := range stmt.Impl.Methods {
			if _, fieldConflict := structType.Fields[method.Name.Value]; fieldConflict {
				sa.error(fmt.Sprintf("member '%s' already declared on %s", method.Name.Value, stmt.Name.Value))
				continue
			}
			if _, duplicate := structType.Methods[method.Name.Value]; duplicate {
				sa.error(fmt.Sprintf("method '%s' already declared on %s", method.Name.Value, stmt.Name.Value))
				continue
			}
			structType.Methods[method.Name.Value] = sa.structFunctionSignature(method)
		}
	}

	for _, function := range stmt.TypeFunctions {
		sa.analyzeStructFunction(function, structType, false)
	}
	if stmt.Impl != nil {
		for _, method := range stmt.Impl.Methods {
			sa.analyzeStructFunction(method, structType, true)
		}
	}

	if len(sa.errors) > startErrors {
		return types.InvalidType
	}
	return structType
}

func (sa *SemanticAnalyzer) structFunctionSignature(function *ast.FunctionStatement) types.MemberDefinition {
	parameters := make([]types.Type, len(function.Parameters))
	for i, parameter := range function.Parameters {
		parameter.Type = sa.resolveType(parameter.Type)
		if types.IsInvalid(parameter.Type) {
			sa.error(fmt.Sprintf("unknown type for parameter %s of %s", parameter.Name.Value, function.Name.Value))
		}
		parameters[i] = parameter.Type
	}

	returns := make([]types.Type, len(function.ReturnTypes))
	for i, returnType := range function.ReturnTypes {
		returnType.Type = sa.resolveType(returnType.Type)
		if types.IsInvalid(returnType.Type) {
			sa.error(fmt.Sprintf("unknown return type for %s", function.Name.Value))
		}
		returns[i] = returnType.Type
	}

	return types.MemberDefinition{
		Name:        function.Name.Value,
		Kind:        types.MemberMethod,
		Parameters:  parameters,
		ReturnTypes: returns,
		Public:      function.Public,
	}
}

func (sa *SemanticAnalyzer) analyzeStructFunction(
	function *ast.FunctionStatement,
	structType *types.Struct,
	bindSelf bool,
) {
	sa.enterScope()
	defer sa.exitScope()

	if bindSelf {
		sa.current.Set("self", &BasicSymbol{name: "self", typ: structType})
	}
	for _, parameter := range function.Parameters {
		sa.current.Set(parameter.Name.Value, &BasicSymbol{name: parameter.Name.Value, typ: parameter.Type})
	}

	returnTypes := make([]types.Type, len(function.ReturnTypes))
	for i, returnType := range function.ReturnTypes {
		returnTypes[i] = returnType.Type
	}
	previousReturnTypes := sa.returnTypes
	previousStruct := sa.currentStruct
	sa.returnTypes = returnTypes
	sa.currentStruct = structType
	defer func() {
		sa.returnTypes = previousReturnTypes
		sa.currentStruct = previousStruct
	}()

	bodyResult := sa.analyzeBlock(function.Body)
	if len(returnTypes) > 0 && !bodyResult.Returns {
		sa.error(fmt.Sprintf("function '%s' declares return types but does not return on all paths", function.Name.Value))
	}
}

func (sa *SemanticAnalyzer) analyzeStructLiteral(literal *ast.StructLiteral) types.Type {
	startErrors := len(sa.errors)
	symbol, ok := sa.current.Get(literal.Name.Value)
	if !ok {
		sa.error(fmt.Sprintf("unknown struct type: %s", literal.Name.Value))
		return types.InvalidType
	}
	structSymbol, ok := symbol.(*StructSymbol)
	if !ok {
		sa.error(fmt.Sprintf("%s is not a struct type", literal.Name.Value))
		return types.InvalidType
	}
	structType := structSymbol.typ

	provided := make(map[string]struct{}, len(literal.Fields))
	for _, field := range literal.Fields {
		if _, duplicate := provided[field.Name.Value]; duplicate {
			sa.error(fmt.Sprintf("field %s supplied more than once", field.Name.Value))
			continue
		}
		provided[field.Name.Value] = struct{}{}

		definition, exists := structType.Fields[field.Name.Value]
		if !exists {
			sa.error(fmt.Sprintf("type %s has no field %s", structType.Name(), field.Name.Value))
			continue
		}
		actualType := sa.analyzeExpression(field.Value)
		if err := sa.requireAssignable(definition.Type, actualType, field.Value); err != nil {
			sa.error(fmt.Sprintf("field %s.%s: %s", structType.Name(), field.Name.Value, err.Error()))
		}
	}

	for name, field := range structType.Fields {
		if _, supplied := provided[name]; !supplied && !field.HasDefault {
			sa.error(fmt.Sprintf("missing required field %s.%s", structType.Name(), name))
		}
	}

	if len(sa.errors) > startErrors {
		return types.InvalidType
	}
	return structType
}

func (sa *SemanticAnalyzer) resolveType(typ types.Type) types.Type {
	switch typed := typ.(type) {
	case nil:
		return typ
	case *types.Named:
		symbol, found := sa.current.Get(typed.CustomName)
		if !found {
			return types.InvalidType
		}
		structSymbol, ok := symbol.(*StructSymbol)
		if !ok {
			return types.InvalidType
		}
		typed.Underlying = structSymbol.typ
		return typed
	case *types.Array:
		elementType := sa.resolveType(typed.ElementType())
		if types.IsInvalid(elementType) {
			return types.InvalidType
		}
		return types.NewArray(elementType, typed.Length())
	case *types.Slice:
		elementType := sa.resolveType(typed.ElementType())
		if types.IsInvalid(elementType) {
			return types.InvalidType
		}
		return types.NewSlice(elementType)
	default:
		return typ
	}
}

func (sa *SemanticAnalyzer) predeclareStruct(stmt *ast.StructStatement) *types.Struct {
	if _, alreadyHandled := sa.structDeclarations[stmt]; alreadyHandled {
		return sa.structDeclarations[stmt]
	}
	if sa.current.ExistsInCurrentScope(stmt.Name.Value) {
		sa.error(fmt.Sprintf("type '%s' already declared in this scope", stmt.Name.Value))
		sa.structDeclarations[stmt] = nil
		return nil
	}

	structType := &types.Struct{
		TypeName:      stmt.Name.Value,
		Fields:        make(map[string]types.StructField, len(stmt.Fields)),
		TypeFunctions: make(map[string]types.MemberDefinition, len(stmt.TypeFunctions)),
		Methods:       make(map[string]types.MemberDefinition),
	}
	sa.current.Set(stmt.Name.Value, &StructSymbol{name: stmt.Name.Value, typ: structType})
	sa.structDeclarations[stmt] = structType
	return structType
}
