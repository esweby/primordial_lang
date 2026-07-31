package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (sa *SemanticAnalyzer) resolveMember(exp *ast.MemberExpression) (types.MemberDefinition, bool) {
	receiverType := sa.analyzeExpression(exp.Receiver)

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

	return member, true
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

		if !types.IsAssignable(expected, actual) {
			sa.error(fmt.Sprintf(
				"argument %d to %s: expected %s, got %s",
				i,
				member.Name,
				expected.Name(),
				actual.Name(),
			))
			return types.InvalidType
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