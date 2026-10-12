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
		return &object.Continue{}
	}

	return nil
}

// evalValue evaluates an expression used as a value. A loop used as a value
// produces the value it breaks with.
func (e *Evaluator) evalValue(expr ast.Expression, env *object.Environment) object.Object {
	if forLoop, ok := expr.(*ast.ForLoop); ok {
		return e.evalForLoop(forLoop, env, true)
	}
	return e.eval(expr, env)
}

func (e *Evaluator) evalDeclareStatement(stmt *ast.DeclareStatement, env *object.Environment) object.Object {
	value := e.evalValue(stmt.Value, env)
	if isError(value) {
		return value
	}
	env.Set(stmt.Name.Value, value)
	return nil
}

func (e *Evaluator) evalAssignStatement(stmt *ast.AssignStatement, env *object.Environment) object.Object {
	switch target := stmt.Target.(type) {
	case *ast.MemberExpression:
		return e.evalMemberAssignment(target, stmt.Value, env)
	case *ast.IndexExpression:
		return e.evalIndexAssignment(target, stmt.Value, env)
	}

	value := e.evalValue(stmt.Value, env)
	if isError(value) {
		return value
	}
	if _, ok := env.Assign(stmt.Name.Value, value); !ok {
		return internalError("assignment to undeclared variable %s", stmt.Name.Value)
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
		return internalError("tuple declaration received %s", value.Type())
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
		return internalError("tuple assignment received %s", value.Type())
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
		value := e.eval(v, env)
		if isError(value) {
			return value
		}
		values = append(values, value)
	}

	return &object.ReturnValue{Value: values}
}

// evalStructLiteral builds a struct from the supplied fields, evaluating the
// declared default for any field that was not supplied.
func (e *Evaluator) evalStructLiteral(literal *ast.StructLiteral, env *object.Environment) object.Object {
	definitionObject, _ := env.Get(literal.Name.Value)
	definition, ok := definitionObject.(*object.StructDefinition)
	if !ok {
		return internalError("%s is not a struct definition", literal.Name.Value)
	}

	provided := make(map[string]ast.Expression, len(literal.Fields))
	for _, field := range literal.Fields {
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
		value := e.eval(expression, evaluationEnv)
		if isError(value) {
			return value
		}
		fields[declared.Name.Value] = value
	}

	return &object.Struct{Name: literal.Name.Value, Definition: definition, Fields: fields}
}

func (e *Evaluator) evalMapLiteral(m *ast.MapLiteral, env *object.Environment) object.Object {
	mapType, ok := m.Type.(*types.Map)
	if !ok {
		return internalError("map literal has type %s", m.Type.Name())
	}

	pairs := make(map[object.HashKey]object.Object, len(m.Pairs))
	originalKeys := make(map[object.HashKey]object.Object, len(m.Pairs))
	keys := make([]object.HashKey, 0, len(m.Pairs))

	for _, pair := range m.Pairs {
		key := e.eval(pair.Key, env)
		if isError(key) {
			return key
		}
		hashable, ok := key.(object.Hashable)
		if !ok {
			return internalError("map key %s is not hashable", key.Type())
		}

		value := e.eval(pair.Value, env)
		if isError(value) {
			return value
		}

		hk := hashable.HashKey()
		pairs[hk] = value
		originalKeys[hk] = key
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
		return internalError("index assignment into %s", lhs.Type())
	}

	index := e.eval(target.Index, env)
	if isError(index) {
		return index
	}
	hashable, ok := index.(object.Hashable)
	if !ok {
		return internalError("map key %s is not hashable", index.Type())
	}

	value := e.eval(valueExpression, env)
	if isError(value) {
		return value
	}

	hk := hashable.HashKey()
	if _, exists := m.Pairs[hk]; !exists {
		m.Keys = append(m.Keys, hk)
		m.OriginalKeys[hk] = index
	}
	m.Pairs[hk] = value

	return m
}

func (e *Evaluator) evalBreakStatement(node *ast.BreakStatement, env *object.Environment) object.Object {
	var label string
	if node.Label != nil {
		label = node.Label.Value
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
