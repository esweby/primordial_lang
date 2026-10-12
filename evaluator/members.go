package evaluator

import (
	"math/big"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/types"
)

func (e *Evaluator) evalMemberProperty(
	exp *ast.MemberExpression,
	env *object.Environment,
) object.Object {
	receiver := e.eval(exp.Receiver, env)
	if isError(receiver) {
		return receiver
	}

	switch receiver := receiver.(type) {
	case *object.Array:
		if exp.Name.Value == "length" {
			return newIntegerObject(big.NewInt(int64(len(receiver.Elements))), types.IntType)
		}
	case *object.Slice:
		if exp.Name.Value == "length" {
			return newIntegerObject(big.NewInt(int64(len(receiver.Elements))), types.IntType)
		}
	case *object.Struct:
		if value, ok := receiver.Fields[exp.Name.Value]; ok {
			return value
		}
	}

	return internalError("member %s on %s", exp.Name.Value, receiver.Type())
}

func (e *Evaluator) evalMemberCall(
	exp *ast.MemberExpression,
	arguments []ast.Expression,
	env *object.Environment,
) object.Object {
	receiver := e.eval(exp.Receiver, env)
	if isError(receiver) {
		return receiver
	}

	args := e.evalExpressions(arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	switch receiver := receiver.(type) {
	case *object.Array:
		if exp.Name.Value == "toSlice" {
			return &object.Slice{Elements: append([]object.Object(nil), receiver.Elements...)}
		}

	case *object.Slice:
		if exp.Name.Value == "append" {
			receiver.Elements = append(receiver.Elements, args[0])
			return nil
		}

	case *object.StructDefinition:
		if function := e.findStructFunction(receiver.Declaration.TypeFunctions, exp.Name.Value); function != nil {
			functionEnv := object.NewEnclosedEnvironment(receiver.Env)
			return e.applyFunction(newStructFunction(function, functionEnv), args)
		}

	case *object.Struct:
		if receiver.Definition.Declaration.Impl != nil {
			if method := e.findStructFunction(receiver.Definition.Declaration.Impl.Methods, exp.Name.Value); method != nil {
				methodEnv := object.NewEnclosedEnvironment(receiver.Definition.Env)
				methodEnv.Set("self", receiver)
				return e.applyFunction(newStructFunction(method, methodEnv), args)
			}
		}
	}

	return internalError("method %s on %s", exp.Name.Value, receiver.Type())
}

func (e *Evaluator) evalMemberAssignment(
	target *ast.MemberExpression,
	valueExpression ast.Expression,
	env *object.Environment,
) object.Object {
	receiverObject := e.eval(target.Receiver, env)
	if isError(receiverObject) {
		return receiverObject
	}
	receiver, ok := receiverObject.(*object.Struct)
	if !ok {
		return internalError("member assignment on %s", receiverObject.Type())
	}

	value := e.eval(valueExpression, env)
	if isError(value) {
		return value
	}
	receiver.Fields[target.Name.Value] = value
	return nil
}

func (e *Evaluator) findStructFunction(functions []*ast.FunctionStatement, name string) *ast.FunctionStatement {
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
