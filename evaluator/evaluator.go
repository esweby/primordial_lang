package evaluator

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/object"
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

		env.Set(node.Name.Value, value)
		return nil
	case *ast.TupleDeclareStatement:
		return evalTupleDeclaration(node, env)
	case *ast.AssignStatement:
		value := Eval(node.Value, env)
		if isError(value) {
			return value
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
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)
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
		return &object.Integer{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalInfixExpression(node.Operator, left, right)
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

func applyFunction(fn object.Object, args []object.Object) object.Object {
	if fn == nil {
		return newError("attempted to call missing function value")
	}

	switch fn := fn.(type) {
	case *object.Function:
		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)
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
		return Eval(ie.Body, env)
	} else if ie.Else != nil {
		return Eval(ie.Else, env)
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

func evalPrefixExpression(op string, r object.Object) object.Object {
	switch op {
	case "!":
		return evalBangOperatorExpression(r)
	case "-":
		return evalMinusPrefixOperatorExpression(r)
	default:
		return newError("unknown operator: %s %s", op, r.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
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

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := getIntegerValue(left)
	rightVal := getIntegerValue(right)

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value

	return &object.String{Value:leftVal + rightVal}
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

func evalMinusPrefixOperatorExpression(expr object.Object) object.Object {
	integer, ok := expr.(*object.Integer)
	if !ok {
		return newError("unknown operator: -%s", expr.Type())
	}

	return &object.Integer{Value: -integer.Value}
}

// Helper functions
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}

	return FALSE
}

func getIntegerValue(o object.Object) int64 {
	return o.(*object.Integer).Value
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
