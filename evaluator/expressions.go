package evaluator

import (
	"math/big"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/types"
)

func (e *Evaluator) evalExpression(exprNode ast.Node, env *object.Environment) object.Object {
	switch node := exprNode.(type) {
	case *ast.IndexExpression:
		left := e.eval(node.Left, env)
		if isError(left) {
			return left
		}

		index := e.eval(node.Index, env)
		if isError(index) {
			return index
		}

		return e.evalIndexExpression(left, index)
	case *ast.StructLiteral:
		return e.evalStructLiteral(node, env)
	case *ast.MapLiteral:
		return e.evalMapLiteral(node, env)
	case *ast.ArrayLiteral:
		return e.evalArrayLiteral(node, env)
	case *ast.SliceLiteral:
		elements := e.evalTypedExpressions(node.Elements, node.Type, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}

		return &object.Slice{Elements: elements}

	case *ast.FunctionLiteral:
		return &object.Function{
			Name:        "",
			Parameters:  node.Parameters,
			ReturnTypes: node.ReturnTypes,
			Body:        node.Body,
			Env:         env,
		}
	case *ast.CallExpression:
		return e.evalCallExpression(node, env)

	case *ast.MemberExpression:
		return e.evalMemberProperty(node, env)
	case *ast.BlockExpression:
		return e.evalBlock(node, env)
	case *ast.IfExpression:
		return e.evalIfExpression(node, env)

	// Expressions
	case *ast.Identifier:
		return e.evalIdentifier(node, env)
	case *ast.IntegerLiteral:
		integerType := node.GetResolvedType()
		if integerType == nil {
			integerType = types.UntypedIntegerType
		}
		return newIntegerObject(node.Value, integerType)
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return e.nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := e.eval(node.Right, env)
		if isError(right) {
			return right
		}
		return e.evalPrefixExpression(node.Operator, right, node.GetResolvedType())
	case *ast.InfixExpression:
		left := e.eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := e.eval(node.Right, env)
		if isError(right) {
			return right
		}
		return e.evalInfixExpression(node.Operator, left, right, node.GetResolvedType())
	}

	return nil
}

func (e *Evaluator) evalCallExpression(call *ast.CallExpression, env *object.Environment) object.Object {
	if member, ok := call.Function.(*ast.MemberExpression); ok {
		return e.evalMemberCall(member, call.Arguments, env)
	}
	if identifier, ok := call.Function.(*ast.Identifier); ok {
		if target, isBuiltinType := types.GetBuiltin(identifier.Value); isBuiltinType && types.IsInteger(target) {
			return e.evalIntegerConversion(call, target, env)
		}
	}

	function := e.eval(call.Function, env)
	if isError(function) {
		return function
	}

	args := e.evalExpressions(call.Arguments, env)
	if len(args) == 1 && isError(args[0]) {
		return args[0]
	}

	return e.applyFunction(function, args)
}

func (e *Evaluator) evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return e.evalArrayIndexExpression(left, index)
	case left.Type() == object.SLICE_OBJ && index.Type() == object.INTEGER_OBJ:
		return e.evalSliceIndexExpression(left, index)
	case left.Type() == object.MAP_OBJ:
		return e.evalMapIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

func (e *Evaluator) evalArrayLiteral(
	arr *ast.ArrayLiteral,
	env *object.Environment,
) object.Object {
	elements := e.evalTypedExpressions(arr.Elements, arr.Type, env)
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

func (e *Evaluator) evalBlock(block *ast.BlockExpression, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range block.Statements {
		result = e.eval(stmt, env)
		if result == nil {
			continue
		}
		switch result.(type) {
		case *object.ReturnValue, *object.Error, *object.Break, *object.Continue:
			return result
		}
	}
	return result
}

func (e *Evaluator) evalExpressions(args []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, expr := range args {
		var evaluated object.Object
		if forLoop, ok := expr.(*ast.ForLoop); ok {
			evaluated = e.evalForLoop(forLoop, env, true)
		} else {
			evaluated = e.eval(expr, env)
		}
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

func (e *Evaluator) evalArrayIndexExpression(array, index object.Object) object.Object {
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

func (e *Evaluator) evalSliceIndexExpression(array, index object.Object) object.Object {
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

func (e *Evaluator) evalMapIndexExpression(m, index object.Object) object.Object {
	mo := m.(*object.Map)

	coercedKey, err := e.coerceRuntimeArgument(index, mo.MapType.Key)
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

func (e *Evaluator) applyFunction(fn object.Object, args []object.Object) object.Object {
	if fn == nil {
		return newError("attempted to call missing function value")
	}

	switch fn := fn.(type) {
	case *object.Function:
		if len(args) != len(fn.Parameters) {
			return newError("wrong number of arguments: expected %d, got %d", len(fn.Parameters), len(args))
		}
		for i, parameter := range fn.Parameters {
			coerced, err := e.coerceRuntimeArgument(args[i], parameter.Type)
			if err != nil {
				return newError("argument %d: %s", i, err.Error())
			}
			args[i] = coerced
		}
		extendedEnv := e.extendFunctionEnv(fn, args)
		evaluated := e.eval(fn.Body, extendedEnv)
		if isError(evaluated) {
			return evaluated
		}
		if _, ok := evaluated.(*object.Break); ok {
			return evaluated
		}
		if _, ok := evaluated.(*object.Continue); ok {
			return evaluated
		}
		return e.coerceFunctionResult(fn, e.unwrapReturnValue(evaluated))
	case *object.Builtin:
		return fn.Fn(args...)
	}

	return newError("not a function: %s", fn.Type())
}

func (e *Evaluator) extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Set(param.Name.Value, args[paramIdx])
	}

	return env
}

func (e *Evaluator) unwrapReturnValue(obj object.Object) object.Object {
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

func (e *Evaluator) evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := e.eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}
	if e.isTruthy(condition) {
		res := e.eval(ie.Body, object.NewEnclosedEnvironment(env))
		if res == nil {
			return nil
		}
		// propagate break/continue/error immediately
		if _, ok := res.(*object.Break); ok {
			return res
		}
		if _, ok := res.(*object.Continue); ok {
			return res
		}
		return res
	} else if ie.Else != nil {
		res := e.eval(ie.Else, object.NewEnclosedEnvironment(env))
		if res == nil {
			return nil
		}
		if _, ok := res.(*object.Break); ok {
			return res
		}
		if _, ok := res.(*object.Continue); ok {
			return res
		}
		return res
	}
	return nil
}

func (e *Evaluator) evalIdentifier(i *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(i.Value); ok {
		return val

	}

	if val, ok := builtins[i.Value]; ok {
		return val
	}

	return newError("%s", "identifier not found: "+i.Value)
}

func (e *Evaluator) evalForLoop(node *ast.ForLoop, env *object.Environment, expectsValue bool) object.Object {
	e.loopDepth++
	if node.Label != nil {
		e.loopLabels = append(e.loopLabels, node.Label.Value)
	}
	defer func() {
		e.loopDepth--
		if node.Label != nil {
			e.loopLabels = e.loopLabels[:len(e.loopLabels)-1]
		}
	}()

	loopEnv := object.NewEnclosedEnvironment(env)
	var result object.Object
	var hasResult bool

	matchesBreak := func(label string) bool {
		if label == "" {
			return true
		}
		if node.Label == nil {
			return false
		}
		return label == node.Label.Value
	}

	switch ctrl := node.Controller.(type) {
	case *ast.Infinite:
		for {
			bodyResult := e.eval(node.Body, loopEnv)
			if isError(bodyResult) {
				return bodyResult
			}
			if br, ok := bodyResult.(*object.Break); ok {
				if matchesBreak(br.Label) {
					if br.Value != nil {
						if expectsValue {
							return br.Value
						}
						result = br.Value
						hasResult = true
					}
					goto exitLoop
				} else {
					return bodyResult
				}
			}
			if _, ok := bodyResult.(*object.Continue); ok {
				continue
			}
		}

	case *ast.While:
		for {
			cond := e.eval(ctrl.Condition, loopEnv)
			if isError(cond) {
				return cond
			}
			if !e.isTruthy(cond) {
				break
			}
			bodyResult := e.eval(node.Body, loopEnv)
			if isError(bodyResult) {
				return bodyResult
			}
			if br, ok := bodyResult.(*object.Break); ok {
				if matchesBreak(br.Label) {
					if br.Value != nil {
						if expectsValue {
							return br.Value
						}
						result = br.Value
						hasResult = true
					}
					goto exitLoop
				} else {
					return bodyResult
				}
			}
			if _, ok := bodyResult.(*object.Continue); ok {
				continue
			}
		}

	case *ast.Constructed:
		e.eval(ctrl.Initializer, loopEnv)
		for {
			cond := e.eval(ctrl.Condition, loopEnv)
			if isError(cond) {
				return cond
			}
			if !e.isTruthy(cond) {
				break
			}
			bodyResult := e.eval(node.Body, loopEnv)
			if isError(bodyResult) {
				return bodyResult
			}
			if br, ok := bodyResult.(*object.Break); ok {
				if matchesBreak(br.Label) {
					if br.Value != nil {
						if expectsValue {
							return br.Value
						}
						result = br.Value
						hasResult = true
					}
					goto exitLoop
				} else {
					return bodyResult
				}
			}
			if _, ok := bodyResult.(*object.Continue); ok {
				// Execute iterator and continue
				e.eval(ctrl.Iterator, loopEnv)
				continue
			}
			// Normal: execute iterator
			e.eval(ctrl.Iterator, loopEnv)
		}

	case *ast.Range:
		iterableObj := e.eval(ctrl.Iterable, loopEnv)
		if isError(iterableObj) {
			return iterableObj
		}
		elements := e.iterableToElements(iterableObj)
		for _, elem := range elements {
			if len(ctrl.Variables) == 1 {
				loopEnv.Set(ctrl.Variables[0].Value, elem)
			} else if len(ctrl.Variables) == 2 {
				if tuple, ok := elem.(*object.Tuple); ok && len(tuple.Elements) == 2 {
					loopEnv.Set(ctrl.Variables[0].Value, tuple.Elements[0])
					loopEnv.Set(ctrl.Variables[1].Value, tuple.Elements[1])
				}
			}
			bodyResult := e.eval(node.Body, loopEnv)
			if isError(bodyResult) {
				return bodyResult
			}
			if br, ok := bodyResult.(*object.Break); ok {
				if matchesBreak(br.Label) {
					if br.Value != nil {
						if expectsValue {
							return br.Value
						}
						result = br.Value
						hasResult = true
					}
					goto exitLoop
				} else {
					return bodyResult
				}
			}
			if _, ok := bodyResult.(*object.Continue); ok {
				continue
			}
		}
	}

exitLoop:
	if expectsValue && !hasResult {
		return object.VOID
	}
	if hasResult {
		return result
	}
	return object.VOID
}

func (e *Evaluator) iterableToElements(obj object.Object) []object.Object {
	switch o := obj.(type) {
	case *object.Array:
		elements := make([]object.Object, len(o.Elements))
		for i, val := range o.Elements {
			elements[i] = &object.Tuple{
				Elements: []object.Object{newIntegerObject(big.NewInt(int64(i)), types.Int64Type), val},
			}
		}
		return elements
	case *object.Slice:
		elements := make([]object.Object, len(o.Elements))
		for i, val := range o.Elements {
			elements[i] = &object.Tuple{
				Elements: []object.Object{newIntegerObject(big.NewInt(int64(i)), types.Int64Type), val},
			}
		}
		return elements
	case *object.Map:
		// To iterate, we need to produce (key, value) tuples.
		// Since we only have HashKey, we can't reconstruct the original key object.
		// We could store original key objects in a separate map: OriginalKeys map[HashKey]Object
		// Add that field to object.Map and fill it in evalMapLiteral.
		// Then here we can retrieve the original key object.
		// We'll implement that now.

		// We'll assume object.Map has OriginalKeys map[HashKey]Object
		elements := make([]object.Object, 0, len(o.Keys))
		for _, hk := range o.Keys {
			keyObj := o.OriginalKeys[hk] // need to add this field
			valueObj := o.Pairs[hk]
			elements = append(elements, &object.Tuple{
				Elements: []object.Object{keyObj, valueObj},
			})
		}
		return elements
	case *object.String:
		runes := []rune(o.Value)
		elements := make([]object.Object, len(runes))
		for i, r := range runes {
			elements[i] = &object.Tuple{
				Elements: []object.Object{
					newIntegerObject(big.NewInt(int64(i)), types.Int64Type),
					&object.String{Value: string(r)},
				},
			}
		}
		return elements
	default:
		return []object.Object{}
	}
}

func (e *Evaluator) evalTypedExpressions(expressions []ast.Expression, target types.Type, env *object.Environment) []object.Object {
	result := make([]object.Object, 0, len(expressions))
	for _, expression := range expressions {
		value := e.eval(expression, env)
		if isError(value) {
			return []object.Object{value}
		}
		coerced, err := e.coerceRuntimeArgument(value, target)
		if err != nil {
			return []object.Object{newError("%s", err.Error())}
		}
		result = append(result, coerced)
	}
	return result
}

func (e *Evaluator) evalPrefixExpression(op string, r object.Object, resultType types.Type) object.Object {
	switch op {
	case "!":
		return e.evalBangOperatorExpression(r)
	case "-":
		return e.evalMinusPrefixOperatorExpression(r, resultType)
	default:
		return newError("unknown operator: %s %s", op, r.Type())
	}
}

func (e *Evaluator) evalInfixExpression(operator string, left, right object.Object, resultType types.Type) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalIntegerInfixExpression(operator, left, right, resultType)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return e.evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return e.nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return e.nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalIntegerInfixExpression(operator string, left, right object.Object, resultType types.Type) object.Object {
	leftInteger := left.(*object.Integer)
	rightInteger := right.(*object.Integer)

	// Coerce untyped integer to match the other operand's concrete type
	if types.IsUntypedInteger(leftInteger.IntegerType) && !types.IsUntypedInteger(rightInteger.IntegerType) {
		coerced, err := e.coerceRuntimeArgument(left, rightInteger.IntegerType)
		if err != nil {
			return newError("integer coercion: %s", err.Error())
		}
		leftInteger = coerced.(*object.Integer)
	} else if types.IsUntypedInteger(rightInteger.IntegerType) && !types.IsUntypedInteger(leftInteger.IntegerType) {
		coerced, err := e.coerceRuntimeArgument(right, leftInteger.IntegerType)
		if err != nil {
			return newError("integer coercion: %s", err.Error())
		}
		rightInteger = coerced.(*object.Integer)
	}

	if !types.IsTypesEqual(leftInteger.IntegerType, rightInteger.IntegerType) {
		return newError("integer type mismatch: %s and %s", leftInteger.IntegerType.Name(), rightInteger.IntegerType.Name())
	}

	leftVal := leftInteger.Value
	rightVal := rightInteger.Value
	if resultType == nil {
		resultType = leftInteger.IntegerType
	}

	if resultType == nil {
		resultType = leftInteger.IntegerType
	}

	switch operator {
	case "+":
		return e.checkedIntegerResult(new(big.Int).Add(leftVal, rightVal), resultType)
	case "-":
		return e.checkedIntegerResult(new(big.Int).Sub(leftVal, rightVal), resultType)
	case "*":
		return e.checkedIntegerResult(new(big.Int).Mul(leftVal, rightVal), resultType)
	case "/":
		if rightVal.Sign() == 0 {
			return newError("division by zero")
		}
		return e.checkedIntegerResult(new(big.Int).Quo(leftVal, rightVal), resultType)
	case "<":
		return e.nativeBoolToBooleanObject(leftVal.Cmp(rightVal) < 0)
	case ">":
		return e.nativeBoolToBooleanObject(leftVal.Cmp(rightVal) > 0)
	case "<=":
		return e.nativeBoolToBooleanObject(leftVal.Cmp(rightVal) <= 0)
	case ">=":
		return e.nativeBoolToBooleanObject(leftVal.Cmp(rightVal) >= 0)
	case "==":
		return e.nativeBoolToBooleanObject(leftVal.Cmp(rightVal) == 0)
	case "!=":
		return e.nativeBoolToBooleanObject(leftVal.Cmp(rightVal) != 0)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftVal + rightVal}
	case "==":
		return e.nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return e.nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func (e *Evaluator) evalIntegerConversion(call *ast.CallExpression, target types.Type, env *object.Environment) object.Object {
	if len(call.Arguments) != 1 {
		return newError("integer conversion to %s expects 1 argument, got %d", target.Name(), len(call.Arguments))
	}
	value := e.eval(call.Arguments[0], env)
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

func (e *Evaluator) evalBangOperatorExpression(expr object.Object) object.Object {
	switch expr {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	default:
		return FALSE
	}
}

func (e *Evaluator) evalMinusPrefixOperatorExpression(expr object.Object, resultType types.Type) object.Object {
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
	return e.checkedIntegerResult(new(big.Int).Neg(integer.Value), resultType)
}