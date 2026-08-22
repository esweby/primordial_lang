package evaluator

import (
	"fmt"
	"math/big"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/types"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node.Statements, env)
	case *ast.DeclareStatement:
		value := Eval(node.Value, env)
		if isError(value) {
			return value
		}
		if node.GetType() != nil {
			coerced, err := coerceRuntimeArgument(value, node.GetType())
			if err != nil {
				return newError("declaration %s: %s", node.Name.Value, err.Error())
			}
			value = coerced
		}

		env.Set(node.Name.Value, value)
		return nil
	case *ast.StructStatement:
		env.Set(node.Name.Value, &object.StructDefinition{Declaration: node, Env: env})
		return nil
	case *ast.StructLiteral:
		return evalStructLiteral(node, env)
	case *ast.MapLiteral:
		return evalMapLiteral(node, env)
	case *ast.ArrayLiteral:
		return evalArrayLiteral(node, env)
	case *ast.SliceLiteral:
		elements := evalTypedExpressions(node.Elements, node.Type, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		return &object.Slice{Elements: elements}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}

		return evalIndexExpression(left, index)
	case *ast.TupleDeclareStatement:
		return evalTupleDeclaration(node, env)
	case *ast.AssignStatement:
		if node.Name == nil {
			target, ok := node.Target.(*ast.MemberExpression)
			if !ok {
				return newError("invalid assignment target")
			}
			return evalMemberAssignment(target, node.Value, env)
		}
		value := Eval(node.Value, env)
		if isError(value) {
			return value
		}
		if current, found := env.Get(node.Name.Value); found {
			if integer, ok := current.(*object.Integer); ok {
				coerced, err := coerceRuntimeArgument(value, integer.IntegerType)
				if err != nil {
					return newError("assignment to %s: %s", node.Name.Value, err.Error())
				}
				value = coerced
			}
		}
		if _, ok := env.Assign(node.Name.Value, value); !ok {
			return newError("identifier not found: %s", node.Name.Value)
		}
		return nil
	case *ast.TupleAssignStatement:
		return evalTupleAssignment(node, env)
	case *ast.FunctionLiteral:
		return &object.Function{
			Name:        "",
			Parameters:  node.Parameters,
			ReturnTypes: node.ReturnTypes,
			Body:        node.Body,
			Env:         env,
		}
	case *ast.FunctionStatement:
		fn := &object.Function{
			Name:        node.Name.Value,
			Parameters:  node.Parameters,
			ReturnTypes: node.ReturnTypes,
			Body:        node.Body,
			Env:         env,
		}

		env.Set(node.Name.Value, fn)
		return fn
	case *ast.CallExpression:
		if member, ok := node.Function.(*ast.MemberExpression); ok {
			return evalMemberCall(member, node.Arguments, env)
		}
		if identifier, ok := node.Function.(*ast.Identifier); ok {
			if target, isBuiltinType := types.GetBuiltin(identifier.Value); isBuiltinType && types.IsInteger(target) {
				return evalIntegerConversion(node, target, env)
			}
		}

		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
	case *ast.MemberExpression:
		return evalMemberProperty(node, env)
	case *ast.BlockExpression:
		return evalBlock(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)
	case *ast.ReturnStatement:
		return evalReturnStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)

	// Expressions
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.IntegerLiteral:
		integerType := node.GetResolvedType()
		if integerType == nil {
			integerType = types.UntypedIntegerType
		}
		return newIntegerObject(node.Value, integerType)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right, node.GetResolvedType())
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right, node.GetResolvedType())
	}

	return nil
}

func evalTupleDeclaration(stmt *ast.TupleDeclareStatement, env *object.Environment) object.Object {
	value := Eval(stmt.Value, env)
	if isError(value) {
		return value
	}

	tuple, ok := value.(*object.Tuple)
	if !ok {
		return newError("tuple declaration requires a tuple value, got %s", value.Type())
	}
	if len(stmt.Names) != len(tuple.Elements) {
		return newError("tuple declaration arity mismatch: expected %d names, got %d",
			len(tuple.Elements), len(stmt.Names))
	}

	for i, name := range stmt.Names {
		if name.Value != "_" {
			env.Set(name.Value, tuple.Elements[i])
		}
	}

	return nil
}

func evalTupleAssignment(stmt *ast.TupleAssignStatement, env *object.Environment) object.Object {
	value := Eval(stmt.Value, env)
	if isError(value) {
		return value
	}

	tuple, ok := value.(*object.Tuple)
	if !ok {
		return newError("tuple assignment requires a tuple value, got %s", value.Type())
	}
	if len(stmt.Names) != len(tuple.Elements) {
		return newError("tuple assignment arity mismatch: expected %d names, got %d",
			len(tuple.Elements), len(stmt.Names))
	}

	// Validate every target before changing any binding so assignment is atomic.
	for _, name := range stmt.Names {
		if name.Value == "_" {
			continue
		}
		if _, found := env.Get(name.Value); !found {
			return newError("identifier not found: %s", name.Value)
		}
	}

	for i, name := range stmt.Names {
		if name.Value != "_" {
			env.Assign(name.Value, tuple.Elements[i])
		}
	}

	return nil
}

func evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range stmts {
		result = Eval(statement, env)

		switch result.(type) {
		case *object.ReturnValue:
			return result
		case *object.Error:
			return result
		}
	}

	return result
}

func evalArrayLiteral(
	arr *ast.ArrayLiteral,
	env *object.Environment,
) object.Object {
	elements := evalTypedExpressions(arr.Elements, arr.Type, env)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}

	for len(elements) < arr.Size {
		neutral, ok := neutralObject(arr.Type)
		if !ok {
			return newError(
				"internal error: no neutral value for type %s",
				arr.Type.Name(),
			)
		}

		elements = append(elements, neutral)
	}

	return &object.Array{Elements: elements}
}

func evalBlock(block *ast.BlockExpression, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result == nil {
			continue
		}

		switch result.(type) {
		case *object.ReturnValue:
			return result
		case *object.Error:
			return result
		}
	}

	return result
}

func evalExpressions(args []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range args {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}

		result = append(result, evaluated)
	}

	return result
}

func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.SLICE_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalSliceIndexExpression(left, index)
	case left.Type() == object.MAP_OBJ:
		return evalMapIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	indexValue := index.(*object.Integer).Value
	if !indexValue.IsInt64() {
		return newError("array index is outside the supported range: %s", indexValue.String())
	}
	idx := indexValue.Int64()
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return newError("array index out of bounds: %d", idx)
	}

	return arrayObject.Elements[idx]
}

func evalSliceIndexExpression(array, index object.Object) object.Object {
	slice := array.(*object.Slice)
	indexValue := index.(*object.Integer).Value
	if !indexValue.IsInt64() {
		return newError("slice index is outside the supported range: %s", indexValue.String())
	}
	idx := indexValue.Int64()
	max := int64(len(slice.Elements) - 1)

	if idx < 0 || idx > max {
		return newError("array index out of bounds: %d", idx)
	}

	return slice.Elements[idx]
}

func evalMapIndexExpression(m, index object.Object) object.Object {
	mo := m.(*object.Map)

	coercedKey, err := coerceRuntimeArgument(index, mo.MapType.Key)
	if err != nil {
		return newError("map index: %s", err.Error())
	}

	hashable, ok := coercedKey.(object.Hashable)
	if !ok {
		return newError("unhashable type as map key: %s", coercedKey.Type())
	}

	value, ok := mo.Pairs[hashable.HashKey()]
	if !ok {
		return newError("key not found in map: %s", hashable.HashKey())
	}

	return value
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	if fn == nil {
		return newError("attempted to call missing function value")
	}

	switch fn := fn.(type) {
	case *object.Function:
		if len(args) != len(fn.Parameters) {
			return newError("wrong number of arguments: expected %d, got %d", len(fn.Parameters), len(args))
		}
		for i, parameter := range fn.Parameters {
			coerced, err := coerceRuntimeArgument(args[i], parameter.Type)
			if err != nil {
				return newError("argument %d: %s", i, err.Error())
			}
			args[i] = coerced
		}
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		if isError(evaluated) {
			return evaluated
		}
		return coerceFunctionResult(fn, unwrapReturnValue(evaluated))
	case *object.Builtin:
		return fn.Fn(args...)
	}

	return newError("not a function: %s", fn.Type())
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Name.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	returnValue, ok := obj.(*object.ReturnValue)
	if !ok {
		return obj
	}

	switch len(returnValue.Value) {
	case 0:
		return nil
	case 1:
		return returnValue.Value[0]
	default:
		return &object.Tuple{
			Elements: returnValue.Value,
		}
	}
}

func evalReturnStatement(rs *ast.ReturnStatement, env *object.Environment) object.Object {
	values := []object.Object{}

	for _, v := range rs.ReturnValues {
		values = append(values, Eval(v, env))
	}

	return &object.ReturnValue{Value: values}
}

func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(ie.Body, object.NewEnclosedEnvironment(env))
	} else if ie.Else != nil {
		return Eval(ie.Else, object.NewEnclosedEnvironment(env))
	}

	return nil
}

func evalIdentifier(i *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(i.Value); ok {
		return val

	}

	if val, ok := builtins[i.Value]; ok {
		return val
	}

	return newError("%s", "identifier not found: "+i.Value)
}

func evalPrefixExpression(op string, r object.Object, resultType types.Type) object.Object {
	switch op {
	case "!":
		return evalBangOperatorExpression(r)
	case "-":
		return evalMinusPrefixOperatorExpression(r, resultType)
	default:
		return newError("unknown operator: %s %s", op, r.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object, resultType types.Type) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right, resultType)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object, resultType types.Type) object.Object {
	leftInteger := left.(*object.Integer)
	rightInteger := right.(*object.Integer)
	leftVal := leftInteger.Value
	rightVal := rightInteger.Value

	if !types.IsTypesEqual(leftInteger.IntegerType, rightInteger.IntegerType) {
		return newError("integer type mismatch: %s and %s", leftInteger.IntegerType.Name(), rightInteger.IntegerType.Name())
	}
	if resultType == nil {
		resultType = leftInteger.IntegerType
	}

	switch operator {
	case "+":
		return checkedIntegerResult(new(big.Int).Add(leftVal, rightVal), resultType)
	case "-":
		return checkedIntegerResult(new(big.Int).Sub(leftVal, rightVal), resultType)
	case "*":
		return checkedIntegerResult(new(big.Int).Mul(leftVal, rightVal), resultType)
	case "/":
		if rightVal.Sign() == 0 {
			return newError("division by zero")
		}
		return checkedIntegerResult(new(big.Int).Quo(leftVal, rightVal), resultType)
	case "<":
		return nativeBoolToBooleanObject(leftVal.Cmp(rightVal) < 0)
	case ">":
		return nativeBoolToBooleanObject(leftVal.Cmp(rightVal) > 0)
	case "<=":
		return nativeBoolToBooleanObject(leftVal.Cmp(rightVal) <= 0)
	case ">=":
		return nativeBoolToBooleanObject(leftVal.Cmp(rightVal) >= 0)
	case "==":
		return nativeBoolToBooleanObject(leftVal.Cmp(rightVal) == 0)
	case "!=":
		return nativeBoolToBooleanObject(leftVal.Cmp(rightVal) != 0)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStructLiteral(literal *ast.StructLiteral, env *object.Environment) object.Object {
	definitionObject, ok := env.Get(literal.Name.Value)
	if !ok {
		return newError("unknown struct type: %s", literal.Name.Value)
	}
	definition, ok := definitionObject.(*object.StructDefinition)
	if !ok {
		return newError("%s is not a struct type", literal.Name.Value)
	}

	declaredFields := make(map[string]*ast.StructField, len(definition.Declaration.Fields))
	for _, field := range definition.Declaration.Fields {
		declaredFields[field.Name.Value] = field
	}

	provided := make(map[string]ast.Expression, len(literal.Fields))
	for _, field := range literal.Fields {
		if _, exists := declaredFields[field.Name.Value]; !exists {
			return newError("type %s has no field %s", literal.Name.Value, field.Name.Value)
		}
		if _, duplicate := provided[field.Name.Value]; duplicate {
			return newError("field %s supplied more than once", field.Name.Value)
		}
		provided[field.Name.Value] = field.Value
	}

	fields := make(map[string]object.Object, len(definition.Declaration.Fields))
	for _, declared := range definition.Declaration.Fields {
		expression, supplied := provided[declared.Name.Value]
		evaluationEnv := env
		if !supplied {
			expression = declared.Value
			evaluationEnv = definition.Env
		}
		if expression == nil {
			return newError("missing required field %s.%s", literal.Name.Value, declared.Name.Value)
		}
		value := Eval(expression, evaluationEnv)
		if isError(value) {
			return value
		}
		coerced, err := coerceRuntimeArgument(value, declared.Type)
		if err != nil {
			return newError("field %s.%s: %s", literal.Name.Value, declared.Name.Value, err.Error())
		}
		fields[declared.Name.Value] = coerced
	}

	return &object.Struct{Name: literal.Name.Value, Definition: definition, Fields: fields}
}

func evalMapLiteral(m *ast.MapLiteral, env *object.Environment) object.Object {
    mapType, ok := m.Type.(*types.Map)
    if !ok {
        return newError("invalid map type: %T", m.Type)
    }

    pairs := make(map[object.HashKey]object.Object, len(m.Pairs))

    for _, pair := range m.Pairs {
        // Evaluate key expression to get an object
        keyObj := Eval(pair.Key, env)
        if isError(keyObj) {
            return keyObj
        }

        // Coerce key to the declared key type (optional but recommended)
        coercedKey, err := coerceRuntimeArgument(keyObj, mapType.Key)
        if err != nil {
            return newError("map key %s: %s", pair.Key.String(), err.Error())
        }

        // Compute hash key from the coerced key object
        hashKey, ok := coercedKey.(object.Hashable)
        if !ok {
            return newError("coercedKey is not object.Hashabkle")
        }

        // Evaluate value expression
        valueObj := Eval(pair.Value, env)
        if isError(valueObj) {
            return valueObj
        }

        // Coerce value to the declared value type
        coercedValue, err := coerceRuntimeArgument(valueObj, mapType.Value)
        if err != nil {
            return newError("map value for key %s: %s", pair.Key.String(), err.Error())
        }

        pairs[hashKey.HashKey()] = coercedValue
    }

    return &object.Map{
		MapType: m.Type.(*types.Map),
		Pairs: pairs,
	}
}

func evalIntegerConversion(call *ast.CallExpression, target types.Type, env *object.Environment) object.Object {
	if len(call.Arguments) != 1 {
		return newError("integer conversion to %s expects 1 argument, got %d", target.Name(), len(call.Arguments))
	}
	value := Eval(call.Arguments[0], env)
	if isError(value) {
		return value
	}
	integer, ok := value.(*object.Integer)
	if !ok {
		return newError("cannot convert %s to %s", value.Type(), target.Name())
	}
	concrete := target.(*types.Integer)
	if !concrete.CanRepresent(integer.Value) {
		return newError("integer conversion overflow: %s is not representable as %s", integer.Value.String(), target.Name())
	}
	return newIntegerObject(integer.Value, target)
}

func evalBangOperatorExpression(expr object.Object) object.Object {
	switch expr {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(expr object.Object, resultType types.Type) object.Object {
	integer, ok := expr.(*object.Integer)
	if !ok {
		return newError("unknown operator: -%s", expr.Type())
	}

	if resultType == nil {
		resultType = integer.IntegerType
	}
	if integerType, ok := resultType.(*types.Integer); ok && !integerType.Signed() {
		return newError("cannot negate unsigned integer %s", resultType.Name())
	}
	return checkedIntegerResult(new(big.Int).Neg(integer.Value), resultType)
}

// Helper functions
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}

	return FALSE
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
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

func newIntegerObject(value *big.Int, integerType types.Type) *object.Integer {
	return &object.Integer{Value: new(big.Int).Set(value), IntegerType: integerType}
}

func checkedIntegerResult(value *big.Int, integerType types.Type) object.Object {
	if concrete, ok := integerType.(*types.Integer); ok && !concrete.CanRepresent(value) {
		return newError("integer overflow: %s is not representable as %s", value.String(), concrete.Name())
	}
	return newIntegerObject(value, integerType)
}

func coerceRuntimeArgument(argument object.Object, target types.Type) (object.Object, error) {
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

func evalTypedExpressions(expressions []ast.Expression, target types.Type, env *object.Environment) []object.Object {
	result := make([]object.Object, 0, len(expressions))
	for _, expression := range expressions {
		value := Eval(expression, env)
		if isError(value) {
			return []object.Object{value}
		}
		coerced, err := coerceRuntimeArgument(value, target)
		if err != nil {
			return []object.Object{newError("%s", err.Error())}
		}
		result = append(result, coerced)
	}
	return result
}

func coerceFunctionResult(function *object.Function, result object.Object) object.Object {
	if result == nil || len(function.ReturnTypes) == 0 {
		return result
	}
	if len(function.ReturnTypes) == 1 {
		coerced, err := coerceRuntimeArgument(result, function.ReturnTypes[0].Type)
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
		coerced, err := coerceRuntimeArgument(tuple.Elements[i], returnType.Type)
		if err != nil {
			return newError("return value %d: %s", i, err.Error())
		}
		tuple.Elements[i] = coerced
	}
	return tuple
}
