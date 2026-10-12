package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (a *Analyzer) analyzeStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		a.analyzeExpressionStatement(s)
	case *ast.DeclareStatement:
		a.analyzeDeclareStatement(s)
	case *ast.TupleDeclareStatement:
		a.analyzeTupleDeclareStatement(s)
	case *ast.AssignStatement:
		a.analyzeAssignStatement(s)
	case *ast.TupleAssignStatement:
		a.analyzeTupleAssignStatement(s)
	case *ast.ReturnStatement:
		a.analyzeReturnStatement(s)
	case *ast.FunctionStatement:
		a.analyzeFunction(s.ReturnTypes, s.Body, fmt.Sprintf("function '%s'", s.Name.Value), s)
	case *ast.StructStatement:
		a.analyzeStructStatement(s)
	case *ast.BreakStatement:
		a.analyzeBreakStatement(s)
	case *ast.ContinueStatement:
	}
}

func (a *Analyzer) analyzeStatements(stmts []ast.Statement) {
	for _, stmt := range stmts {
		a.analyzeStatement(stmt)
	}
}

// analyzeExpressionStatement analyzes an expression whose value is unused. If
// and for are statements here, so they do not need to produce a value.
func (a *Analyzer) analyzeExpressionStatement(stmt *ast.ExpressionStatement) {
	switch e := stmt.Expression.(type) {
	case *ast.IfExpression:
		a.analyzeIfExpression(e, false)
	case *ast.ForLoop:
		a.analyzeForLoop(e, false)
	default:
		a.analyzeExpression(e)
	}
}

func (a *Analyzer) analyzeDeclareStatement(stmt *ast.DeclareStatement) {
	a.analyzeExpression(stmt.Value)
	if stmt.Type != nil && !stmt.Inferred {
		a.checkAssignable(stmt, stmt.Type, typeOf(stmt.Value), "declaration type mismatch")
	}
}

func (a *Analyzer) analyzeTupleDeclareStatement(stmt *ast.TupleDeclareStatement) {
	a.analyzeExpression(stmt.Value)
	a.checkTuple(stmt, stmt.Value, len(stmt.Names), "declaration")
}

// checkTuple reports a value that is not a tuple of the given arity.
func (a *Analyzer) checkTuple(node ast.Node, value ast.Expression, names int, kind string) ([]types.Type, bool) {
	valueType := typeOf(value)
	if types.IsInvalid(valueType) {
		return nil, false
	}

	elements, ok := types.UnwrapTuple(valueType)
	if !ok {
		a.errorAt(node, "tuple %s requires a tuple value, got %s", kind, valueType.Name())
		return nil, false
	}
	if len(elements) != names {
		a.errorAt(node, "tuple %s arity mismatch: expected %d names, got %d", kind, len(elements), names)
		return nil, false
	}
	return elements, true
}

func (a *Analyzer) analyzeAssignStatement(stmt *ast.AssignStatement) {
	if stmt.Name != nil {
		if !a.checkAssignableName(stmt.Name) {
			a.analyzeExpression(stmt.Value)
			return
		}
		a.analyzeExpression(stmt.Value)
		a.checkAssignable(stmt, typeOf(stmt.Name), typeOf(stmt.Value), "assignment type mismatch")
		return
	}

	switch target := stmt.Target.(type) {
	case *ast.MemberExpression:
		member, found := a.lookupMember(target)
		a.analyzeExpression(stmt.Value)
		if !found {
			return
		}
		if member.Kind != types.MemberProperty || len(member.ReturnTypes) != 1 {
			a.errorAt(target, "cannot assign to method: %s", member.Name)
			return
		}
		a.checkAssignable(stmt, member.ReturnTypes[0], typeOf(stmt.Value), "assignment type mismatch")
	case *ast.IndexExpression:
		a.analyzeExpression(target)
		a.analyzeExpression(stmt.Value)
		a.checkAssignable(stmt, typeOf(target), typeOf(stmt.Value), "assignment type mismatch")
	default:
		a.analyzeExpression(stmt.Value)
		a.errorAt(stmt, "invalid target assignment")
	}
}

// checkAssignableName reports a name that is not a mutable variable.
func (a *Analyzer) checkAssignableName(name *ast.Identifier) bool {
	sym := symbolOf(name)
	if sym == nil {
		return false
	}
	isVariable, mutable := variable(sym)
	if !isVariable {
		a.errorAt(name, "cannot assign to non-variable: %s", name.Value)
		return false
	}
	if !mutable {
		a.errorAt(name, "cannot assign to immutable variable: %s", name.Value)
		return false
	}
	return true
}

func (a *Analyzer) analyzeTupleAssignStatement(stmt *ast.TupleAssignStatement) {
	a.analyzeExpression(stmt.Value)
	elements, ok := a.checkTuple(stmt, stmt.Value, len(stmt.Names), "assignment")

	for i, name := range stmt.Names {
		if name.Value == "_" || !a.checkAssignableName(name) || !ok {
			continue
		}
		a.checkAssignable(name, typeOf(name), elements[i], fmt.Sprintf("tuple assignment value %d", i))
	}
}

func (a *Analyzer) analyzeReturnStatement(stmt *ast.ReturnStatement) {
	for _, value := range stmt.ReturnValues {
		a.analyzeExpression(value)
	}

	var returnTypes []types.Type
	if a.function != nil {
		returnTypes = a.function.returnTypes
	}

	if len(returnTypes) == 0 {
		if len(stmt.ReturnValues) > 0 {
			a.errorAt(stmt, "unexpected return with values in void function")
		}
		return
	}

	if len(stmt.ReturnValues) != len(returnTypes) {
		a.errorAt(stmt, "wrong number of return values: expected %d, got %d", len(returnTypes), len(stmt.ReturnValues))
		return
	}

	for i, value := range stmt.ReturnValues {
		a.checkAssignable(value, returnTypes[i], typeOf(value), fmt.Sprintf("return value %d", i))
	}
}

// analyzeFunction analyzes a function body with its return types. Loops do
// not cross a function boundary.
func (a *Analyzer) analyzeFunction(returns []*ast.ReturnType, body *ast.BlockExpression, name string, node ast.Node) {
	returnTypes := make([]types.Type, len(returns))
	for i, ret := range returns {
		returnTypes[i] = ret.Type
	}

	previousFunction, previousLoops := a.function, a.loops
	a.function, a.loops = &functionContext{returnTypes: returnTypes}, nil
	defer func() {
		a.function, a.loops = previousFunction, previousLoops
	}()

	if body != nil {
		a.analyzeStatements(body.Statements)
	}

	if len(returnTypes) > 0 && !blockReturns(body) {
		a.errorAt(node, "%s declares return types but does not return on all paths", name)
	}
}

func (a *Analyzer) analyzeBreakStatement(stmt *ast.BreakStatement) {
	a.analyzeExpression(stmt.Value)

	for i := len(a.loops) - 1; i >= 0; i-- {
		loop := a.loops[i]
		if stmt.Label == nil || (loop.loop.Label != nil && loop.loop.Label.Value == stmt.Label.Value) {
			loop.breaks = append(loop.breaks, stmt)
			return
		}
	}
}

func (a *Analyzer) analyzeStructStatement(stmt *ast.StructStatement) {
	structType, _ := stmt.Name.GetResolvedType().(*types.Struct)
	name := stmt.Name.Value

	fields := make(map[string]struct{}, len(stmt.Fields))
	for _, field := range stmt.Fields {
		if _, duplicate := fields[field.Name.Value]; duplicate {
			a.errorAt(field.Name, "field '%s' already declared on %s", field.Name.Value, name)
			continue
		}
		fields[field.Name.Value] = struct{}{}

		if field.Type == nil && field.Value == nil {
			a.errorAt(field.Name, "field %s.%s requires a type or default value", name, field.Name.Value)
			continue
		}
		if field.Value != nil {
			a.analyzeExpression(field.Value)
			if field.Type != nil && !field.Inferred {
				a.checkAssignable(field.Value, field.Type, typeOf(field.Value),
					fmt.Sprintf("default value for %s.%s", name, field.Name.Value))
			}
		}
	}

	typeFunctions := make(map[string]struct{}, len(stmt.TypeFunctions))
	for _, function := range stmt.TypeFunctions {
		if _, duplicate := typeFunctions[function.Name.Value]; duplicate {
			a.errorAt(function.Name, "type function '%s' already declared on %s", function.Name.Value, name)
		}
		typeFunctions[function.Name.Value] = struct{}{}
	}

	if stmt.Impl != nil {
		methods := make(map[string]struct{}, len(stmt.Impl.Methods))
		for _, method := range stmt.Impl.Methods {
			if _, conflict := fields[method.Name.Value]; conflict {
				a.errorAt(method.Name, "member '%s' already declared on %s", method.Name.Value, name)
			} else if _, duplicate := methods[method.Name.Value]; duplicate {
				a.errorAt(method.Name, "method '%s' already declared on %s", method.Name.Value, name)
			}
			methods[method.Name.Value] = struct{}{}
		}
	}

	previousStruct := a.currentStruct
	a.currentStruct = structType
	defer func() { a.currentStruct = previousStruct }()

	for _, function := range stmt.TypeFunctions {
		a.analyzeFunction(function.ReturnTypes, function.Body, fmt.Sprintf("function '%s'", function.Name.Value), function)
	}
	if stmt.Impl != nil {
		for _, method := range stmt.Impl.Methods {
			a.analyzeFunction(method.ReturnTypes, method.Body, fmt.Sprintf("function '%s'", method.Name.Value), method)
		}
	}
}
