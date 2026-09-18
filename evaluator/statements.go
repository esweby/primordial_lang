package evaluator

import (
	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/types"
)

func (e *Evaluator) evalStatement(stmtNode ast.Node, env *object.Environment) object.Object {
	switch node := stmtNode.(type) {
	case *ast.DeclareStatement:
		return e.evalDeclareStatement(node, env)
	case *ast.StructStatement:
		env.Set(node.Name.Value, &object.StructDefinition{Declaration: node, Env: env})
		return nil
	case *ast.AssignStatement:
		return e.evalAssignStatement(node, env)
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
	case *ast.TupleDeclareStatement:
		return e.evalTupleDeclaration(node, env)
	case *ast.TupleAssignStatement:
		return e.evalTupleAssignment(node, env)
	case *ast.ExpressionStatement:
		if forLoop, ok := node.Expression.(*ast.ForLoop); ok {
			return e.evalForLoop(forLoop, env, false)
		}
		return e.eval(node.Expression, env)
	case *ast.ReturnStatement:
		return e.evalReturnStatement(node, env)
	case *ast.BreakStatement:
		return e.evalBreakStatement(node, env)
	case *ast.ContinueStatement:
		return e.evalContinueStatement(node, env)
	} 

	return nil
}

func (e *Evaluator) evalDeclareStatement(stmt *ast.DeclareStatement, env *object.Environment) object.Object {
	var value object.Object
	if forLoop, ok := stmt.Value.(*ast.ForLoop); ok {
		value = e.evalForLoop(forLoop, env, true)
	} else {
		value = e.eval(stmt.Value, env)
	}
	if isError(value) {
		return value
	}
	if stmt.GetType() != nil {
		coerced, err := e.coerceRuntimeArgument(value, stmt.GetType())
		if err != nil {
			return newError("declaration %s: %s", stmt.Name.Value, err.Error())
		}
		value = coerced
	}
	env.Set(stmt.Name.Value, value)
	return nil
}

func (e *Evaluator) evalAssignStatement(stmt *ast.AssignStatement, env *object.Environment) object.Object {
	if stmt.Name == nil {
		// handle member/index assignment (existing logic)
		target, ok := stmt.Target.(*ast.MemberExpression)
		if !ok {
			if target, ok := stmt.Target.(*ast.IndexExpression); ok {
				return e.evalIndexAssignment(target, stmt.Value, env)
			}
			return newError("invalid assignment target")
		}
		return e.evalMemberAssignment(target, stmt.Value, env)
	}
	var value object.Object
	if forLoop, ok := stmt.Value.(*ast.ForLoop); ok {
		value = e.evalForLoop(forLoop, env, true)
	} else {
		value = e.eval(stmt.Value, env)
	}
	if isError(value) {
		return value
	}
	if current, found := env.Get(stmt.Name.Value); found {
		if integer, ok := current.(*object.Integer); ok {
			coerced, err := e.coerceRuntimeArgument(value, integer.IntegerType)
			if err != nil {
				return newError("assignment to %s: %s", stmt.Name.Value, err.Error())
			}
			value = coerced
		}
	}
	if _, ok := env.Assign(stmt.Name.Value, value); !ok {
		return newError("identifier not found: %s", stmt.Name.Value)
	}
	return nil
}

func (e *Evaluator) evalTupleDeclaration(stmt *ast.TupleDeclareStatement, env *object.Environment) object.Object {
	value := e.eval(stmt.Value, env)
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

func (e *Evaluator) evalTupleAssignment(stmt *ast.TupleAssignStatement, env *object.Environment) object.Object {
	value := e.eval(stmt.Value, env)
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

func (e *Evaluator) evalReturnStatement(rs *ast.ReturnStatement, env *object.Environment) object.Object {
	values := []object.Object{}

	for _, v := range rs.ReturnValues {
		values = append(values, e.eval(v, env))
	}

	return &object.ReturnValue{Value: values}
}

func (e *Evaluator) evalStructLiteral(literal *ast.StructLiteral, env *object.Environment) object.Object {
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
		value := e.eval(expression, evaluationEnv)
		if isError(value) {
			return value
		}
		coerced, err := e.coerceRuntimeArgument(value, declared.Type)
		if err != nil {
			return newError("field %s.%s: %s", literal.Name.Value, declared.Name.Value, err.Error())
		}
		fields[declared.Name.Value] = coerced
	}

	return &object.Struct{Name: literal.Name.Value, Definition: definition, Fields: fields}
}

func (e *Evaluator) evalMapLiteral(m *ast.MapLiteral, env *object.Environment) object.Object {
	mapType, ok := m.Type.(*types.Map)
	if !ok {
		return newError("invalid map type: %T", m.Type)
	}

	pairs := make(map[object.HashKey]object.Object, len(m.Pairs))
	originalKeys := make(map[object.HashKey]object.Object, len(m.Pairs))
	keys := make([]object.HashKey, 0, len(m.Pairs))

	for _, pair := range m.Pairs {
		// Evaluate key
		keyObj := e.eval(pair.Key, env)
		if isError(keyObj) {
			return keyObj
		}
		// Coerce key to map key type
		coercedKey, err := e.coerceRuntimeArgument(keyObj, mapType.Key)
		if err != nil {
			return newError("map key %s: %s", pair.Key.String(), err.Error())
		}
		hashable, ok := coercedKey.(object.Hashable)
		if !ok {
			return newError("coercedKey is not object.Hashable")
		}
		hk := hashable.HashKey()

		// Evaluate value
		valueObj := e.eval(pair.Value, env)
		if isError(valueObj) {
			return valueObj
		}
		coercedValue, err := e.coerceRuntimeArgument(valueObj, mapType.Value)
		if err != nil {
			return newError("map value for key %s: %s", pair.Key.String(), err.Error())
		}

		pairs[hk] = coercedValue
		originalKeys[hk] = coercedKey // store the coerced key object
		keys = append(keys, hk)
	}

	return &object.Map{
		MapType:      mapType,
		Pairs:        pairs,
		Keys:         keys,
		OriginalKeys: originalKeys,
	}
}

func (e *Evaluator) evalIndexAssignment(
	target *ast.IndexExpression,
	valueExpression ast.Expression,
	env *object.Environment,
) object.Object {
	lhs := e.eval(target.Left, env)
	if isError(lhs) {
		return lhs
	}

	m, ok := lhs.(*object.Map)
	if !ok {
		return newError("cannot assign to non-map object %s", lhs.Type())
	}

	index := e.eval(target.Index, env)
	if isError(index) {
		return index
	}

	coercedKey, err := e.coerceRuntimeArgument(index, m.MapType.Key)
	if err != nil {
		return newError("map key %s: %s", m.MapType.Key.Name(), err.Error())
	}

	hashKey, ok := coercedKey.(object.Hashable)
	if !ok {
		return newError("coercedKey is not object.Hashabkle")
	}

	value := e.eval(valueExpression, env)
	if isError(value) {
		return value
	}

	m.Pairs[hashKey.HashKey()] = value

	return m
}

func (e *Evaluator) evalBreakStatement(node *ast.BreakStatement, env *object.Environment) object.Object {
	if e.loopDepth == 0 {
		return newError("break outside of loop")
	}

	var label string
	if node.Label != nil {
		label = node.Label.Value
		// Check if label exists in the stack
		found := false
		for _, lbl := range e.loopLabels {
			if lbl == label {
				found = true
				break
			}
		}
		if !found {
			return newError("break label '%s' does not match any enclosing loop", label)
		}
	}
	var value object.Object
	if node.Value != nil {
		value = e.eval(node.Value, env)
		if isError(value) {
			return value
		}
	}
	return &object.Break{Label: label, Value: value}
}

func (e *Evaluator) evalContinueStatement(_ *ast.ContinueStatement, _ *object.Environment) object.Object {
	if e.loopDepth == 0 {
		return newError("continue outside of loop")
	}
	return &object.Continue{}
}