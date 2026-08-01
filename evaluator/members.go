package evaluator

import (
	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/object"
)

func evalMemberProperty(
	exp *ast.MemberExpression,
	env *object.Environment,
) object.Object {
	receiver := Eval(exp.Receiver, env)
	if isError(receiver) {
		return receiver
	}

	switch receiver := receiver.(type) {
	case *object.Array:
		switch exp.Name.Value {
		case "length":
			return &object.Integer{
				Value: int64(len(receiver.Elements)),
			}
		}

	case *object.Slice:
		switch exp.Name.Value {
		case "length":
			return &object.Integer{
				Value: int64(len(receiver.Elements)),
			}
		}
	}

	return newError(
		"member %s is not available on %s",
		exp.Name.Value,
		receiver.Type(),
	)
}

func evalMemberCall(
	exp *ast.MemberExpression,
	arguments []ast.Expression,
	env *object.Environment,
) object.Object {
	receiver := Eval(exp.Receiver, env)
	if isError(receiver) {
		return receiver
	}

	args := evalExpressions(arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	switch receiver := receiver.(type) {
	case *object.Array:
		return evalArrayMethod(
			receiver,
			exp.Name.Value,
			args,
		)

	case *object.Slice:
		return evalSliceMethod(
			receiver,
			exp.Name.Value,
			args,
		)

	default:
		return newError(
			"type %s has no methods",
			receiver.Type(),
		)
	}
}

func evalSliceMethod(
	slice *object.Slice,
	name string,
	args []object.Object,
) object.Object {
	switch name {
	case "append":
		if len(args) != 1 {
			return newError(
				"append expects 1 argument, got %d",
				len(args),
			)
		}

		slice.Elements = append(slice.Elements, args[0])
		return nil

	default:
		return newError(
			"slice has no method %s",
			name,
		)
	}
}

func evalArrayMethod(
	array *object.Array,
	name string,
	args []object.Object,
) object.Object {
	switch name {
	case "toSlice":
		if len(args) != 0 {
			return newError(
				"toSlice expects 0 arguments, got %d",
				len(args),
			)
		}

		elements := append(
			[]object.Object(nil),
			array.Elements...,
		)

		return &object.Slice{
			Elements: elements,
		}

	default:
		return newError(
			"array has no method %s",
			name,
		)
	}
}
