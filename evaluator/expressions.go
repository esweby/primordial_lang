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
		elements := e.evalExpressions(node.Elements, env)
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
		if !types.IsInteger(node.GetResolvedType()) {
			return internalError("integer literal %s has no concrete type", node.Value)
		}
		return newIntegerObject(node.Value, node.GetResolvedType())
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return e.nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := e.eval(node.Right, env)
		if isError(right) {
			return right
		}
		return e.evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := e.eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := e.eval(node.Right, env)
		if isError(right) {
			return right
		}
		return e.evalInfixExpression(node.Operator, left, right)
	}

	return nil
}

func (e *Evaluator) evalCallExpression(call *ast.CallExpression, env *object.Environment) object.Object {
	if member, ok := call.Function.(*ast.MemberExpression); ok {
		return e.evalMemberCall(member, call.Arguments, env)
	}
	if identifier, ok := call.Function.(*ast.Identifier); ok && isBuiltin(identifier) {
		if target, isType := types.GetBuiltin(identifier.Value); isType && types.IsInteger(target) {
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
		return internalError("index into %s", left.Type())
	}
}

func (e *Evaluator) evalArrayLiteral(
	arr *ast.ArrayLiteral,
	env *object.Environment,
) object.Object {
	elements := e.evalExpressions(arr.Elements, env)
	if len(elements) == 1 && isError(elements[0]) {
		return elements[0]
	}

	for len(elements) < arr.Size {
		neutral, ok := neutralObject(arr.Type)
		if !ok {
			return internalError("no neutral value for type %s", arr.Type.Name())
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
		evaluated := e.evalValue(expr, env)
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

	hashable, ok := index.(object.Hashable)
	if !ok {
		return internalError("map key %s is not hashable", index.Type())
	}

	value, ok := mo.Pairs[hashable.HashKey()]
	if !ok {
		return newError("key not found in map: %s", hashable.HashKey())
	}

	return value
}

func (e *Evaluator) applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
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
		return e.unwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(args...)
	}

	return internalError("call to a value that is not a function")
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

	return internalError("identifier %s has no value", i.Value)
}

func (e *Evaluator) evalForLoop(node *ast.ForLoop, env *object.Environment, expectsValue bool) object.Object {
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
				Elements: []object.Object{newIntegerObject(big.NewInt(int64(i)), types.IntType), val},
			}
		}
		return elements
	case *object.Slice:
		elements := make([]object.Object, len(o.Elements))
		for i, val := range o.Elements {
			elements[i] = &object.Tuple{
				Elements: []object.Object{newIntegerObject(big.NewInt(int64(i)), types.IntType), val},
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
	default:
		return []object.Object{}
	}
}

func (e *Evaluator) evalPrefixExpression(op string, r object.Object) object.Object {
	switch op {
	case "!":
		return e.nativeBoolToBooleanObject(r != TRUE)
	case "-":
		return e.evalMinusPrefixOperatorExpression(r)
	default:
		return internalError("unknown prefix operator %s", op)
	}
}

func (e *Evaluator) evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return e.evalIntegerInfixExpression(operator, left.(*object.Integer), right.(*object.Integer))
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return e.evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return e.nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return e.nativeBoolToBooleanObject(left != right)
	default:
		return internalError("operator %s on %s and %s", operator, left.Type(), right.Type())
	}
}

// evalIntegerInfixExpression applies an operator to two integers of the same
// type, trapping results that do not fit it.
func (e *Evaluator) evalIntegerInfixExpression(operator string, left, right *object.Integer) object.Object {
	leftVal, rightVal, resultType := left.Value, right.Value, left.IntegerType

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
		return internalError("unknown integer operator %s", operator)
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
		return internalError("unknown string operator %s", operator)
	}
}

// evalIntegerConversion converts an integer to another integer type, trapping
// a value that does not fit.
func (e *Evaluator) evalIntegerConversion(call *ast.CallExpression, target types.Type, env *object.Environment) object.Object {
	value := e.eval(call.Arguments[0], env)
	if isError(value) {
		return value
	}
	integer, ok := value.(*object.Integer)
	if !ok {
		return internalError("conversion of %s to %s", value.Type(), target.Name())
	}
	concrete := target.(*types.Integer)
	if !concrete.CanRepresent(integer.Value) {
		return newError("integer conversion overflow: %s is not representable as %s", integer.Value.String(), target.Name())
	}
	return newIntegerObject(integer.Value, target)
}

func (e *Evaluator) evalMinusPrefixOperatorExpression(expr object.Object) object.Object {
	integer, ok := expr.(*object.Integer)
	if !ok {
		return internalError("negation of %s", expr.Type())
	}
	return e.checkedIntegerResult(new(big.Int).Neg(integer.Value), integer.IntegerType)
}

// isBuiltin reports whether an identifier names a builtin rather than a
// declaration in the program.
func isBuiltin(ident *ast.Identifier) bool {
	return ident.Symbol != nil && ident.Symbol.Declaration() == nil
}
