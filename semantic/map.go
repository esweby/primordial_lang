package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (sa *SemanticAnalyzer) analyzeMapLiteral(exp *ast.MapLiteral) types.Type {
	startErrors := len(sa.errors)

	expType, ok := exp.Type.(*types.Map)
	if !ok {
		sa.error(fmt.Sprintf("%s type is not a map type", exp.Type.Name()))
		return types.InvalidType
	}

	for i, pair := range exp.Pairs {
		key := sa.analyzeExpression(pair.Key)
		if types.IsInvalid(key) {
			sa.error(fmt.Sprintf("pair %d: key is Invalid Type", i))
			continue
		}

		err := sa.requireAssignable(expType.Key, key, pair.Key)
		if err != nil {
			sa.error(
				fmt.Sprintf(
					"pair %d: key (%s) does not equal expected type %s", 
					i,
					key.Name(),
					expType.Key.Name(),
				),
			)
			continue
		}

		value := sa.analyzeExpression(pair.Value)
		if types.IsInvalid(value) {
			sa.error(fmt.Sprintf("pair %d: value is Invalid Type", i))
			continue
		}

		err = sa.requireAssignable(expType.Value, value, pair.Value)
		if err != nil {
			sa.error(
				fmt.Sprintf(
					"pair %d: value (%s) does not equal expected type %s", 
					i,
					value.Name(),
					expType.Value.Name(),
				),
			)
			continue
		}
	}

	if len(sa.errors) > startErrors {
		return types.InvalidType
	}

	return expType
}