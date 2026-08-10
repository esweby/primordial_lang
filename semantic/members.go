package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (sa *SemanticAnalyzer) resolveMember(exp *ast.MemberExpression) (types.MemberDefinition, bool) {
	if identifier, ok := exp.Receiver.(*ast.Identifier); ok {
		if symbol, found := sa.current.Get(identifier.Value); found {
			if structSymbol, isStructType := symbol.(*StructSymbol); isStructType {
				member, exists := structSymbol.typ.LookupTypeMember(exp.Name.Value)
				if !exists {
					sa.error(fmt.Sprintf("type %s has no type function %s", structSymbol.typ.Name(), exp.Name.Value))
					return types.MemberDefinition{}, false
				}
				return member, true
			}
		}
	}

	receiverType := sa.analyzeExpression(exp.Receiver)
	if types.IsUntypedInteger(receiverType) {
		resolved, err := sa.defaultInteger(exp.Receiver, receiverType)
		if err != nil {
			sa.error(err.Error())
			return types.MemberDefinition{}, false
		}
		receiverType = resolved
	}

	if types.IsInvalid(receiverType) {
		return types.MemberDefinition{}, false
	}

	provider, ok := receiverType.(types.MemberProvider)
	if !ok {
		sa.error(fmt.Sprintf(
			"type %s has no members",
			receiverType.Name(),
		))
		return types.MemberDefinition{}, false
	}

	member, ok := provider.LookupMember(exp.Name.Value)
	if !ok {
		sa.error(fmt.Sprintf(
			"type %s has no member %s",
			receiverType.Name(),
			exp.Name.Value,
		))
		return types.MemberDefinition{}, false
	}
	if receiverType.Kind() == types.KindStruct && member.Kind == types.MemberProperty && !member.Public {
		if sa.currentStruct == nil || underlyingStruct(receiverType) != sa.currentStruct {
			sa.error(fmt.Sprintf("field %s.%s is private", receiverType.Name(), exp.Name.Value))
			return types.MemberDefinition{}, false
		}
	}

	return member, true
}

func underlyingStruct(typ types.Type) *types.Struct {
	switch typed := typ.(type) {
	case *types.Struct:
		return typed
	case *types.Named:
		return underlyingStruct(typed.Underlying)
	default:
		return nil
	}
}

func (sa *SemanticAnalyzer) analyzeMemberExpression(exp *ast.MemberExpression) types.Type {
	member, ok := sa.resolveMember(exp)
	if !ok {
		return types.InvalidType
	}

	if member.Kind != types.MemberProperty {
		sa.error(fmt.Sprintf(
			"method %s must be called",
			member.Name,
		))
		return types.InvalidType
	}

	if len(member.ReturnTypes) != 1 {
		sa.error(fmt.Sprintf(
			"internal error: property %s must have one type",
			member.Name,
		))
		return types.InvalidType
	}

	return member.ReturnTypes[0]
}

func (sa *SemanticAnalyzer) analyzeMemberCall(exp *ast.MemberExpression, args []ast.Expression) types.Type {
	member, ok := sa.resolveMember(exp)
	if !ok {
		return types.InvalidType
	}

	if member.Kind != types.MemberMethod {
		sa.error(fmt.Sprintf(
			"property %s is not callable",
			member.Name,
		))
		return types.InvalidType
	}

	if len(args) != len(member.Parameters) {
		sa.error(fmt.Sprintf(
			"%s expects %d arguments, got %d",
			member.Name,
			len(member.Parameters),
			len(args),
		))
		return types.InvalidType
	}

	for i, argument := range args {
		actual := sa.analyzeExpression(argument)
		expected := member.Parameters[i]

		if err := sa.requireAssignable(expected, actual, argument); err != nil {
			sa.error(fmt.Sprintf(
				"argument %d to %s: %s",
				i,
				member.Name,
				err.Error(),
			))
			return types.InvalidType
		}
	}

	if member.MutatesReceiver {
		if identifier, ok := exp.Receiver.(*ast.Identifier); ok {
			symbol, found := sa.current.Get(identifier.Value)
			if found {
				declaration, variable := symbol.(*DeclareSymbol)
				if !variable || !declaration.Mutable() {
					sa.error(fmt.Sprintf("cannot call mutating method %s on immutable variable: %s", member.Name, identifier.Value))
					return types.InvalidType
				}
			}
		}
	}

	switch len(member.ReturnTypes) {
	case 0:
		return types.VoidType
	case 1:
		return member.ReturnTypes[0]
	default:
		return &types.Tuple{Types: member.ReturnTypes}
	}
}
