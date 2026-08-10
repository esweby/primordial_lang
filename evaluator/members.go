package evaluator

import (
	"math/big"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/types"
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
			return newIntegerObject(big.NewInt(int64(len(receiver.Elements))), types.Int64Type)
		}

	case *object.Slice:
		switch exp.Name.Value {
		case "length":
			return newIntegerObject(big.NewInt(int64(len(receiver.Elements))), types.Int64Type)
		}
	case *object.Struct:
		value, ok := receiver.Fields[exp.Name.Value]
		if ok {
			field := findStructField(receiver.Definition, exp.Name.Value)
			if field != nil && !field.Public && !hasStructAccess(receiver, env) {
				return newError("field %s.%s is private", receiver.Name, exp.Name.Value)
			}
			return value
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

	case *object.StructDefinition:
		function := findStructFunction(receiver.Declaration.TypeFunctions, exp.Name.Value)
		if function == nil {
			return newError("struct %s has no type function %s", receiver.Declaration.Name.Value, exp.Name.Value)
		}
		functionEnv := object.NewEnclosedEnvironment(receiver.Env)
		functionEnv.SetStructContext(receiver)
		return applyFunction(newStructFunction(function, functionEnv), args)

	case *object.Struct:
		if receiver.Definition == nil || receiver.Definition.Declaration.Impl == nil {
			return newError("struct %s has no method %s", receiver.Name, exp.Name.Value)
		}
		method := findStructFunction(receiver.Definition.Declaration.Impl.Methods, exp.Name.Value)
		if method == nil {
			return newError("struct %s has no method %s", receiver.Name, exp.Name.Value)
		}
		methodEnv := object.NewEnclosedEnvironment(receiver.Definition.Env)
		methodEnv.Set("self", receiver)
		methodEnv.SetStructContext(receiver.Definition)
		return applyFunction(newStructFunction(method, methodEnv), args)

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

func evalMemberAssignment(
	target *ast.MemberExpression,
	valueExpression ast.Expression,
	env *object.Environment,
) object.Object {
	receiverObject := Eval(target.Receiver, env)
	if isError(receiverObject) {
		return receiverObject
	}
	receiver, ok := receiverObject.(*object.Struct)
	if !ok {
		return newError("cannot assign member %s on %s", target.Name.Value, receiverObject.Type())
	}

	field := findStructField(receiver.Definition, target.Name.Value)
	if field == nil {
		return newError("struct %s has no field %s", receiver.Name, target.Name.Value)
	}
	if !field.Public && !hasStructAccess(receiver, env) {
		return newError("field %s.%s is private", receiver.Name, target.Name.Value)
	}
	if !hasStructAccess(receiver, env) {
		return newError("cannot assign to field outside its struct: %s", target.Name.Value)
	}

	value := Eval(valueExpression, env)
	if isError(value) {
		return value
	}
	coerced, err := coerceRuntimeArgument(value, field.Type)
	if err != nil {
		return newError("assignment to %s.%s: %s", receiver.Name, target.Name.Value, err.Error())
	}
	receiver.Fields[target.Name.Value] = coerced
	return nil
}

func hasStructAccess(receiver *object.Struct, env *object.Environment) bool {
	return env.StructContext() == receiver.Definition
}

func findStructField(definition *object.StructDefinition, name string) *ast.StructField {
	if definition == nil || definition.Declaration == nil {
		return nil
	}
	for _, field := range definition.Declaration.Fields {
		if field.Name.Value == name {
			return field
		}
	}
	return nil
}

func findStructFunction(functions []*ast.FunctionStatement, name string) *ast.FunctionStatement {
	for _, function := range functions {
		if function.Name.Value == name {
			return function
		}
	}
	return nil
}

func newStructFunction(function *ast.FunctionStatement, env *object.Environment) *object.Function {
	return &object.Function{
		Name:        function.Name.Value,
		Parameters:  function.Parameters,
		ReturnTypes: function.ReturnTypes,
		Body:        function.Body,
		Env:         env,
	}
}
