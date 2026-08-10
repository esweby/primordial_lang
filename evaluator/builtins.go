package evaluator

import (
	"math/big"

	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/types"
)

var builtins = map[string]*object.Builtin{
	"len": {
		Fn: func(args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.Array:
				return newIntegerObject(big.NewInt(int64(len(arg.Elements))), types.Int64Type)
			case *object.Slice:
				return newIntegerObject(big.NewInt(int64(len(arg.Elements))), types.Int64Type)
			case *object.String:
				return newIntegerObject(big.NewInt(int64(len(arg.Value))), types.Int64Type)
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
}
