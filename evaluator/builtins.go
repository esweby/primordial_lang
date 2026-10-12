package evaluator

import (
	"math/big"

	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/types"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) object.Object {
			switch arg := args[0].(type) {
			case *object.Array:
				return newIntegerObject(big.NewInt(int64(len(arg.Elements))), types.IntType)
			case *object.Slice:
				return newIntegerObject(big.NewInt(int64(len(arg.Elements))), types.IntType)
			case *object.String:
				return newIntegerObject(big.NewInt(int64(len(arg.Value))), types.IntType)
			default:
				return internalError("len of %s", args[0].Type())
			}
		},
	},
}
