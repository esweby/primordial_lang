package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/resolver"
	"github.com/esweby/primordial_lang/types"
)

func (a *Analyzer) analyzeExpression(expr ast.Expression) {
	switch e := expr.(type) {
	case nil:
	case *ast.Identifier:
		a.analyzeIdentifier(e)
	case *ast.PrefixExpression:
		a.analyzePrefixExpression(e)
	case *ast.InfixExpression:
		a.analyzeInfixExpression(e)
	case *ast.CallExpression:
		a.analyzeCallExpression(e)
	case *ast.MemberExpression:
		if member, found := a.lookupMember(e); found && member.Kind != types.MemberProperty {
			a.errorAt(e, "method %s must be called", member.Name)
		}
	case *ast.IndexExpression:
		a.analyzeIndexExpression(e)
	case *ast.IfExpression:
		a.analyzeIfExpression(e, true)
	case *ast.BlockExpression:
		a.analyzeStatements(e.Statements)
	case *ast.ForLoop:
		a.analyzeForLoop(e, true)
	case *ast.FunctionLiteral:
		a.analyzeFunction(e.ReturnTypes, e.Body, "function literal", e)
	case *ast.ArrayLiteral:
		a.analyzeArrayLiteral(e)
	case *ast.SliceLiteral:
		if e.Type == nil {
			a.errorAt(e, "slice literal has no element type")
			return
		}
		a.analyzeElements("slice", e.Type, e.Elements)
	case *ast.MapLiteral:
		a.analyzeMapLiteral(e)
	case *ast.StructLiteral:
		a.analyzeStructLiteral(e)
	}
}

// analyzeIdentifier reports names that are not values: types and builtins.
func (a *Analyzer) analyzeIdentifier(ident *ast.Identifier) {
	sym := symbolOf(ident)
	if sym == nil {
		return
	}
	switch sym.Kind {
	case resolver.StructSymbol:
		a.errorAt(ident, "type %s cannot be used as a value", ident.Value)
	case resolver.BuiltinSymbol:
		a.errorAt(ident, "builtin %s must be called", ident.Value)
	}
}

func (a *Analyzer) analyzePrefixExpression(e *ast.PrefixExpression) {
	a.analyzeExpression(e.Right)
	right := typeOf(e.Right)
	if types.IsInvalid(right) {
		return
	}

	switch e.Operator {
	case "-":
		if !types.IsNumeric(right) {
			a.errorAt(e, "operator - requires a number, got %s", right.Name())
		}
	case "!":
		if !types.IsBoolean(right) {
			a.errorAt(e, "operator ! requires a bool, got %s", right.Name())
		}
	default:
		a.errorAt(e, "unknown prefix operator: %s", e.Operator)
	}
}

func (a *Analyzer) analyzeInfixExpression(e *ast.InfixExpression) {
	a.analyzeExpression(e.Left)
	a.analyzeExpression(e.Right)
	left, right := typeOf(e.Left), typeOf(e.Right)
	if types.IsInvalid(left) || types.IsInvalid(right) {
		return
	}
	if types.IsInteger(left) && types.IsInteger(right) {
		return
	}

	switch e.Operator {
	case "+", "-", "*", "/":
		if e.Operator == "+" && types.IsString(left) && types.IsString(right) {
			return
		}
		if !types.IsNumeric(left) || !types.IsTypesEqual(left, right) {
			a.errorAt(e, "mismatched types: %s %s %s", left.Name(), e.Operator, right.Name())
		}
	case "<", "<=", ">", ">=":
		if !types.IsNumeric(left) || !types.IsTypesEqual(left, right) {
			a.errorAt(e, "mismatched types: %s %s %s", left.Name(), e.Operator, right.Name())
		}
	case "==", "!=":
		if !types.IsNumeric(left) && !types.IsBoolean(left) && !types.IsString(left) {
			a.errorAt(e, "type %s cannot be compared", left.Name())
			return
		}
		if !types.IsTypesEqual(left, right) {
			a.errorAt(e, "mismatched types: %s %s %s", left.Name(), e.Operator, right.Name())
		}
	default:
		a.errorAt(e, "unknown infix operator: %s", e.Operator)
	}
}

func (a *Analyzer) analyzeCallExpression(e *ast.CallExpression) {
	for _, arg := range e.Arguments {
		a.analyzeExpression(arg)
	}

	switch callee := e.Function.(type) {
	case *ast.Identifier:
		if sym := symbolOf(callee); sym != nil && sym.Kind == resolver.BuiltinSymbol {
			// Integer conversions are checked by inference.
			if callee.Value == "len" {
				a.analyzeLenCall(e)
			}
			return
		}
	case *ast.MemberExpression:
		a.analyzeMemberCall(e, callee)
		return
	}

	a.analyzeExpression(e.Function)
	calleeType := typeOf(e.Function)
	if types.IsInvalid(calleeType) {
		return
	}
	if calleeType == types.FunctionType {
		a.errorAt(e, "cannot call a value of type function; give it a signature type such as fn(int64): int64")
		return
	}

	signature, ok := underlying(calleeType).(*types.Function)
	if !ok {
		a.errorAt(e, "cannot call non-function %s", calleeType.Name())
		return
	}
	if len(e.Arguments) != len(signature.ParamTypes) {
		a.errorAt(e, "wrong number of arguments: expected %d, got %d", len(signature.ParamTypes), len(e.Arguments))
		return
	}
	for i, arg := range e.Arguments {
		a.checkAssignable(arg, signature.ParamTypes[i], typeOf(arg), fmt.Sprintf("argument %d", i))
	}
}

func (a *Analyzer) analyzeLenCall(e *ast.CallExpression) {
	if len(e.Arguments) != 1 {
		a.errorAt(e, "len expects 1 argument, got %d", len(e.Arguments))
		return
	}

	argument := typeOf(e.Arguments[0])
	switch underlying(argument).(type) {
	case *types.Array, *types.Slice, *types.String:
	default:
		if !types.IsInvalid(argument) {
			a.errorAt(e, "argument to len not supported, got %s", argument.Name())
		}
	}
}

func (a *Analyzer) analyzeMemberCall(e *ast.CallExpression, callee *ast.MemberExpression) {
	member, found := a.lookupMember(callee)
	if !found {
		return
	}
	if member.Kind != types.MemberMethod {
		a.errorAt(callee, "property %s is not callable", member.Name)
		return
	}
	if len(e.Arguments) != len(member.Parameters) {
		a.errorAt(e, "%s expects %d arguments, got %d", member.Name, len(member.Parameters), len(e.Arguments))
		return
	}
	for i, arg := range e.Arguments {
		a.checkAssignable(arg, member.Parameters[i], typeOf(arg), fmt.Sprintf("argument %d to %s", i, member.Name))
	}

	if receiver, ok := callee.Receiver.(*ast.Identifier); ok && member.MutatesReceiver {
		if _, mutable := variable(symbolOf(receiver)); !mutable {
			a.errorAt(e, "cannot call mutating method %s on immutable variable: %s", member.Name, receiver.Value)
		}
	}
}

// lookupMember finds a member on the receiver's type, or a type function when
// the receiver names a struct, reporting members that do not exist or are
// private to another struct.
func (a *Analyzer) lookupMember(e *ast.MemberExpression) (types.MemberDefinition, bool) {
	if ident, ok := e.Receiver.(*ast.Identifier); ok {
		if sym := symbolOf(ident); sym != nil && sym.Kind == resolver.StructSymbol {
			structType, isStruct := typeOf(ident).(*types.Struct)
			if !isStruct {
				return types.MemberDefinition{}, false
			}
			member, exists := structType.LookupTypeMember(e.Name.Value)
			if !exists {
				a.errorAt(e, "type %s has no type function %s", structType.Name(), e.Name.Value)
			}
			return member, exists
		}
	}

	a.analyzeExpression(e.Receiver)
	receiver := typeOf(e.Receiver)
	if types.IsInvalid(receiver) {
		return types.MemberDefinition{}, false
	}

	provider, ok := underlying(receiver).(types.MemberProvider)
	if !ok {
		a.errorAt(e, "type %s has no members", receiver.Name())
		return types.MemberDefinition{}, false
	}
	member, ok := provider.LookupMember(e.Name.Value)
	if !ok {
		a.errorAt(e, "type %s has no member %s", receiver.Name(), e.Name.Value)
		return types.MemberDefinition{}, false
	}

	if structType, isStruct := underlying(receiver).(*types.Struct); isStruct &&
		member.Kind == types.MemberProperty && !member.Public && structType != a.currentStruct {
		a.errorAt(e, "field %s.%s is private", receiver.Name(), e.Name.Value)
		return types.MemberDefinition{}, false
	}

	return member, true
}

func (a *Analyzer) analyzeIndexExpression(e *ast.IndexExpression) {
	a.analyzeExpression(e.Left)
	a.analyzeExpression(e.Index)

	collectionType := typeOf(e.Left)
	if types.IsInvalid(collectionType) {
		return
	}
	collection, ok := underlying(collectionType).(types.Indexable)
	if !ok {
		a.errorAt(e, "type %s cannot be indexed", collectionType.Name())
		return
	}

	index := typeOf(e.Index)
	if !types.IsInvalid(index) && !sameType(collection.IndexType(), index) {
		a.errorAt(e.Index, "index type mismatch: got %s, want %s", index.Name(), collection.IndexType().Name())
	}
}

// branch is what an if branch produces.
type branch struct {
	typ     types.Type
	yields  bool
	returns bool
}

// analyzeIfExpression analyzes an if. As a value it needs an else, and its
// branches must produce values of the same type unless they return.
func (a *Analyzer) analyzeIfExpression(e *ast.IfExpression, asValue bool) branch {
	a.analyzeExpression(e.Condition)
	if condition := typeOf(e.Condition); !types.IsInvalid(condition) && !types.IsBoolean(condition) {
		a.errorAt(e.Condition, "if condition must be boolean, got %s", condition.Name())
	}

	then := a.analyzeBranch(e.Body, asValue)
	var otherwise branch
	hasElse := e.Else != nil
	switch elseBranch := e.Else.(type) {
	case *ast.BlockExpression:
		otherwise = a.analyzeBranch(elseBranch, asValue)
	case *ast.IfExpression:
		otherwise = a.analyzeIfExpression(elseBranch, asValue)
	}

	result := branch{returns: hasElse && then.returns && otherwise.returns}
	if !asValue {
		return result
	}

	switch {
	case !hasElse:
		a.errorAt(e, "if expression used as value requires an else branch")
	case then.returns && otherwise.returns:
		a.errorAt(e, "if expression branches both return, no value produced")
	case then.returns:
		if !otherwise.yields {
			a.errorAt(e, "else branch must yield an expression when then branch returns")
		}
		result.typ, result.yields = otherwise.typ, otherwise.yields
	case otherwise.returns:
		if !then.yields {
			a.errorAt(e, "then branch must yield an expression when else branch returns")
		}
		result.typ, result.yields = then.typ, then.yields
	case !then.yields || !otherwise.yields:
		a.errorAt(e, "if expression branches must end with an expression")
	default:
		if !types.IsInvalid(then.typ) && !types.IsInvalid(otherwise.typ) && !sameType(then.typ, otherwise.typ) {
			a.errorAt(e, "if expression branches have mismatched types: %s vs %s", then.typ.Name(), otherwise.typ.Name())
		}
		result.typ, result.yields = then.typ, true
	}
	return result
}

// analyzeBranch analyzes a branch's statements. As a value, its final
// expression statement is its value.
func (a *Analyzer) analyzeBranch(block *ast.BlockExpression, asValue bool) branch {
	if block == nil {
		return branch{}
	}

	result := branch{returns: blockReturns(block)}
	last := len(block.Statements) - 1
	for i, stmt := range block.Statements {
		if tail, ok := stmt.(*ast.ExpressionStatement); ok && asValue && i == last {
			a.analyzeExpression(tail.Expression)
			result.typ, result.yields = typeOf(tail.Expression), true
			continue
		}
		a.analyzeStatement(stmt)
	}
	return result
}

// analyzeForLoop analyzes a loop. A loop used as a value must break, and its
// breaks must agree on whether they carry a value and on its type.
func (a *Analyzer) analyzeForLoop(loop *ast.ForLoop, asValue bool) {
	switch c := loop.Controller.(type) {
	case *ast.While:
		a.analyzeExpression(c.Condition)
		a.checkCondition(c.Condition, "while")
	case *ast.Constructed:
		a.analyzeStatement(c.Initializer)
		a.analyzeExpression(c.Condition)
		a.checkCondition(c.Condition, "loop")
		a.analyzeStatement(c.Iterator)
	case *ast.Range:
		a.analyzeExpression(c.Iterable)
		iterable := typeOf(c.Iterable)
		switch underlying(iterable).(type) {
		case *types.Array, *types.Slice, *types.Map:
		default:
			if !types.IsInvalid(iterable) {
				a.errorAt(c.Iterable, "range can only iterate over an array, slice, or map, got %s", iterable.Name())
			}
		}
	}

	ctx := &loopContext{loop: loop}
	a.loops = append(a.loops, ctx)
	if loop.Body != nil {
		a.analyzeStatements(loop.Body.Statements)
	}
	a.loops = a.loops[:len(a.loops)-1]

	var valueType types.Type
	withValue, withoutValue := false, false
	for _, brk := range ctx.breaks {
		if brk.Value == nil {
			withoutValue = true
			continue
		}
		withValue = true
		breakType := typeOf(brk.Value)
		if types.IsInvalid(breakType) {
			continue
		}
		if valueType != nil && !sameType(valueType, breakType) {
			a.errorAt(brk, "break value types must match: %s and %s", valueType.Name(), breakType.Name())
		}
		valueType = breakType
	}

	switch {
	case asValue && len(ctx.breaks) == 0:
		a.errorAt(loop, "loop used as expression must have a break to exit")
	case withValue && withoutValue:
		a.errorAt(loop, "mixed break statements: some with values and some without")
	}
}

func (a *Analyzer) checkCondition(condition ast.Expression, kind string) {
	if typ := typeOf(condition); !types.IsInvalid(typ) && !types.IsBoolean(typ) {
		a.errorAt(condition, "%s condition must be boolean, got %s", kind, typ.Name())
	}
}

func (a *Analyzer) analyzeArrayLiteral(e *ast.ArrayLiteral) {
	if e.Type == nil {
		a.errorAt(e, "array literal has no element type")
		return
	}
	if e.Size <= 0 {
		a.errorAt(e, "array length must be greater than zero")
		return
	}
	if len(e.Elements) > e.Size {
		a.errorAt(e, "array length is %d, but %d elements were supplied", e.Size, len(e.Elements))
		return
	}
	a.analyzeElements("array", e.Type, e.Elements)
}

func (a *Analyzer) analyzeElements(collection string, elementType types.Type, elements []ast.Expression) {
	for i, element := range elements {
		a.analyzeExpression(element)
		a.checkAssignable(element, elementType, typeOf(element), fmt.Sprintf("%s element %d", collection, i))
	}
}

func (a *Analyzer) analyzeMapLiteral(e *ast.MapLiteral) {
	mapType, ok := e.Type.(*types.Map)
	if !ok {
		a.errorAt(e, "%s is not a map type", e.Type.Name())
		return
	}

	for i, pair := range e.Pairs {
		a.analyzeExpression(pair.Key)
		a.analyzeExpression(pair.Value)
		a.checkAssignable(pair.Key, mapType.Key, typeOf(pair.Key), fmt.Sprintf("pair %d: key", i))
		a.checkAssignable(pair.Value, mapType.Value, typeOf(pair.Value), fmt.Sprintf("pair %d: value", i))
	}
}

func (a *Analyzer) analyzeStructLiteral(e *ast.StructLiteral) {
	for _, field := range e.Fields {
		a.analyzeExpression(field.Value)
	}

	structType, ok := typeOf(e).(*types.Struct)
	if !ok {
		return
	}

	provided := make(map[string]struct{}, len(e.Fields))
	for _, field := range e.Fields {
		if _, duplicate := provided[field.Name.Value]; duplicate {
			a.errorAt(field.Name, "field %s supplied more than once", field.Name.Value)
			continue
		}
		provided[field.Name.Value] = struct{}{}

		definition, exists := structType.Fields[field.Name.Value]
		if !exists {
			a.errorAt(field.Name, "type %s has no field %s", structType.Name(), field.Name.Value)
			continue
		}
		a.checkAssignable(field.Value, definition.Type, typeOf(field.Value),
			fmt.Sprintf("field %s.%s", structType.Name(), field.Name.Value))
	}

	for name, field := range structType.Fields {
		if _, supplied := provided[name]; !supplied && !field.HasDefault {
			a.errorAt(e, "missing required field %s.%s", structType.Name(), name)
		}
	}
}
