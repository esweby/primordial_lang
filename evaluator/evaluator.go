package evaluator

import (
	"fmt"
	"math/big"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/types"
)

var (
	TRUE       = &object.Boolean{Value: true}
	FALSE      = &object.Boolean{Value: false}
)

type Evaluator struct {
	loopDepth int
	loopLabels []string

	node ast.Node
	env *object.Environment
}

func New(node ast.Node, env *object.Environment) *Evaluator {
	return &Evaluator{
		loopDepth: 0,
		loopLabels: []string{},
		node: node,
		env: env,
	}
}

// In the original monkey language the Evaluator is given the first 
// ast.Node created by the parser and start recursively working its 
// way through the ast. Evaluate will be the public API to start
// this process.
func (e *Evaluator) Evaluate() object.Object {
	program, ok := e.node.(*ast.Program)
	if !ok {
		// no program provided
		return nil
	}
	return e.evalProgram(program.Statements, e.env)
}

func (e *Evaluator) evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range stmts {
		result = e.eval(stmt, env)
		switch result.(type) {
		case *object.ReturnValue, *object.Error, *object.Break, *object.Continue:
			return result
		}
	}
	return result
}

func (e *Evaluator) eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case ast.Statement:
		return e.evalStatement(node, env)
	case ast.Expression:
		return e.evalExpression(node, env)
	default:
		// TODO: Diagnostics
		return nil
	}
}

// Helper functions
func (e *Evaluator) nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}

	return FALSE
}

func (e *Evaluator) isTruthy(obj object.Object) bool {
	switch obj {
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func (e *Evaluator) coerceRuntimeArgument(argument object.Object, target types.Type) (object.Object, error) {
	if argument == nil || target == nil {
		return argument, nil
	}
	integer, isInteger := argument.(*object.Integer)
	targetInteger, targetIsInteger := target.(*types.Integer)
	if isInteger && targetIsInteger {
		if types.IsUntypedInteger(integer.IntegerType) {
			if !targetInteger.CanRepresent(integer.Value) {
				return nil, fmt.Errorf("integer constant %s is not representable as %s", integer.Value.String(), target.Name())
			}
			return newIntegerObject(integer.Value, target), nil
		}
		if !types.IsTypesEqual(integer.IntegerType, target) {
			return nil, fmt.Errorf("expected %s, got %s", target.Name(), integer.IntegerType.Name())
		}
		return argument, nil
	}

	switch target.Kind() {
	case types.KindInteger:
		return nil, fmt.Errorf("expected %s, got %s", target.Name(), argument.Type())
	case types.KindString:
		if _, ok := argument.(*object.String); !ok {
			return nil, fmt.Errorf("expected string, got %s", argument.Type())
		}
	case types.KindBoolean:
		if _, ok := argument.(*object.Boolean); !ok {
			return nil, fmt.Errorf("expected bool, got %s", argument.Type())
		}
	case types.KindStruct:
		value, ok := argument.(*object.Struct)
		if !ok || value.Name != target.Name() {
			return nil, fmt.Errorf("expected %s, got %s", target.Name(), argument.Type())
		}
	}
	return argument, nil
}

func (e *Evaluator) coerceFunctionResult(function *object.Function, result object.Object) object.Object {
	if result == nil || len(function.ReturnTypes) == 0 {
		return result
	}
	if len(function.ReturnTypes) == 1 {
		coerced, err := e.coerceRuntimeArgument(result, function.ReturnTypes[0].Type)
		if err != nil {
			return newError("return value 0: %s", err.Error())
		}
		return coerced
	}
	tuple, ok := result.(*object.Tuple)
	if !ok || len(tuple.Elements) != len(function.ReturnTypes) {
		return newError("return value arity mismatch")
	}
	for i, returnType := range function.ReturnTypes {
		coerced, err := e.coerceRuntimeArgument(tuple.Elements[i], returnType.Type)
		if err != nil {
			return newError("return value %d: %s", i, err.Error())
		}
		tuple.Elements[i] = coerced
	}
	return tuple
}

func newIntegerObject(value *big.Int, integerType types.Type) *object.Integer {
	return &object.Integer{Value: new(big.Int).Set(value), IntegerType: integerType}
}

func (e *Evaluator) checkedIntegerResult(value *big.Int, integerType types.Type) object.Object {
	if concrete, ok := integerType.(*types.Integer); ok && !concrete.CanRepresent(value) {
		return newError("integer overflow: %s is not representable as %s", value.String(), concrete.Name())
	}
	return newIntegerObject(value, integerType)
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.ERROR_OBJ
}

func neutralObject(t types.Type) (object.Object, bool) {
	switch t.Kind() {
	case types.KindInteger:
		return newIntegerObject(big.NewInt(0), t), true
	case types.KindBoolean:
		return &object.Boolean{Value: false}, true
	case types.KindString:
		return &object.String{Value: ""}, true
	default:
		return nil, false
	}
}