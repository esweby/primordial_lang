package inference

import (
	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/resolver"
	"github.com/esweby/primordial_lang/types"
)

// inferValue infers an expression used where target is expected.
func (in *Inferrer) inferValue(expr ast.Expression, target types.Type) types.Type {
	return in.assign(expr, in.inferExpression(expr, target), target)
}

// inferDefault infers an expression with no expected type, defaulting an
// untyped constant to int.
func (in *Inferrer) inferDefault(expr ast.Expression) types.Type {
	return in.defaultUntyped(expr, in.inferExpression(expr, nil))
}

// inferExpression sets and returns the type of expr. The result may be the
// untyped integer type, which the caller converts. expected only guides
// expressions that produce a value from their branches: if, for and blocks.
func (in *Inferrer) inferExpression(expr ast.Expression, expected types.Type) types.Type {
	switch e := expr.(type) {
	case nil:
		return types.InvalidType
	case *ast.IntegerLiteral:
		return setType(e, types.UntypedIntegerType)
	case *ast.StringLiteral:
		return setType(e, types.StringType)
	case *ast.Boolean:
		return setType(e, types.BoolType)
	case *ast.Identifier:
		return setType(e, symbolType(e))
	case *ast.PrefixExpression:
		return setType(e, in.inferPrefixExpression(e))
	case *ast.InfixExpression:
		return setType(e, in.inferInfixExpression(e))
	case *ast.CallExpression:
		return setType(e, in.inferCallExpression(e))
	case *ast.MemberExpression:
		return setType(e, in.inferMemberExpression(e))
	case *ast.IndexExpression:
		return setType(e, in.inferIndexExpression(e))
	case *ast.IfExpression:
		return in.inferIfExpression(e, expected, true).typ
	case *ast.BlockExpression:
		return setType(e, in.inferBlock(e, expected, true).typ)
	case *ast.ForLoop:
		return setType(e, in.inferForLoop(e, expected))
	case *ast.FunctionLiteral:
		return setType(e, in.inferFunctionLiteral(e))
	case *ast.ArrayLiteral:
		in.resolveNamed(e.Type)
		in.inferElements(e.Elements, e.Type)
		return setType(e, types.NewArray(e.Type, e.Size))
	case *ast.SliceLiteral:
		in.resolveNamed(e.Type)
		in.inferElements(e.Elements, e.Type)
		return setType(e, types.NewSlice(e.Type))
	case *ast.MapLiteral:
		return setType(e, in.inferMapLiteral(e))
	case *ast.StructLiteral:
		return setType(e, in.inferStructLiteral(e))
	default:
		return setType(e, types.InvalidType)
	}
}

func (in *Inferrer) inferPrefixExpression(e *ast.PrefixExpression) types.Type {
	right := in.inferExpression(e.Right, nil)

	switch e.Operator {
	case "-":
		if types.IsUntypedInteger(right) {
			return types.UntypedIntegerType
		}
		if integer, ok := right.(*types.Integer); ok && !integer.Signed() {
			in.errorAt(e, "cannot negate unsigned integer %s", right.Name())
			return types.InvalidType
		}
		if types.IsNumeric(right) {
			return right
		}
		return types.InvalidType
	case "!":
		if types.IsBoolean(right) {
			return right
		}
		in.defaultUntyped(e.Right, right)
		return types.InvalidType
	default:
		in.defaultUntyped(e.Right, right)
		return types.InvalidType
	}
}

func (in *Inferrer) inferInfixExpression(e *ast.InfixExpression) types.Type {
	left := in.inferExpression(e.Left, nil)
	right := in.inferExpression(e.Right, nil)

	arithmetic := false
	switch e.Operator {
	case "+", "-", "*", "/":
		arithmetic = true
	case "<", "<=", ">", ">=", "==", "!=":
	default:
		in.defaultUntyped(e.Left, left)
		in.defaultUntyped(e.Right, right)
		return types.InvalidType
	}

	leftUntyped, rightUntyped := types.IsUntypedInteger(left), types.IsUntypedInteger(right)
	var operand types.Type
	switch {
	case leftUntyped && rightUntyped:
		if arithmetic {
			in.checkDivision(e)
			return types.UntypedIntegerType
		}
		operand = in.defaultUntyped(e.Left, left)
		in.defaultUntyped(e.Right, right)
	case leftUntyped:
		operand = right
		in.assign(e.Left, left, right)
	case rightUntyped:
		operand = left
		in.assign(e.Right, right, left)
	default:
		operand = left
		if types.IsInteger(left) && types.IsInteger(right) && !types.IsTypesEqual(left, right) {
			in.errorAt(e, "mismatched types %s and %s", left.Name(), right.Name())
			return types.InvalidType
		}
	}

	if types.IsInteger(operand) && e.Operator == "/" {
		in.checkDivision(e)
	}

	if types.IsInvalid(left) || types.IsInvalid(right) {
		return types.InvalidType
	}
	if !arithmetic {
		return types.BoolType
	}
	if types.IsTypesEqual(left, right) || leftUntyped || rightUntyped {
		return operand
	}
	return types.InvalidType
}

// checkDivision reports integer division by a constant zero.
func (in *Inferrer) checkDivision(e *ast.InfixExpression) {
	if e.Operator != "/" {
		return
	}
	if divisor, ok := constantValue(e.Right); ok && divisor.Sign() == 0 {
		in.errorAt(e, "division by zero")
	}
}

func (in *Inferrer) inferCallExpression(e *ast.CallExpression) types.Type {
	if ident, ok := e.Function.(*ast.Identifier); ok && isBuiltin(ident) {
		setType(ident, types.InvalidType)
		if target, isType := types.GetBuiltin(ident.Value); isType && types.IsInteger(target) {
			return in.inferIntegerConversion(e, target)
		}
		if ident.Value == "len" {
			for _, arg := range e.Arguments {
				in.inferDefault(arg)
			}
			return types.IntType
		}
	}

	var params, returns []types.Type
	callable := false

	if member, ok := e.Function.(*ast.MemberExpression); ok {
		definition, found := in.lookupMember(member)
		if found && definition.Kind == types.MemberMethod {
			params, returns, callable = definition.Parameters, definition.ReturnTypes, true
			setType(member, types.NewFunction(params, returns))
		} else {
			setType(member, types.InvalidType)
		}
	} else {
		callee := in.inferExpression(e.Function, nil)
		if callee == types.FunctionType {
			// A value annotated `function` has no known signature, so untyped
			// constant arguments stay untyped and take the parameter's type
			// when the function is called.
			for _, arg := range e.Arguments {
				in.inferExpression(arg, nil)
			}
			return types.FunctionType
		}
		if signature, ok := underlying(callee).(*types.Function); ok {
			params, returns, callable = signature.ParamTypes, signature.ReturnTypes, true
		}
	}

	in.inferArguments(e.Arguments, params)
	if !callable {
		return types.InvalidType
	}
	return returnType(returns)
}

// inferArguments infers call arguments against the parameters they are passed to.
func (in *Inferrer) inferArguments(args []ast.Expression, params []types.Type) {
	for i, arg := range args {
		if i < len(params) {
			in.inferValue(arg, params[i])
		} else {
			in.inferDefault(arg)
		}
	}
}

// inferIntegerConversion types a conversion such as int32(x). A constant
// argument must fit the target type.
func (in *Inferrer) inferIntegerConversion(e *ast.CallExpression, target types.Type) types.Type {
	if len(e.Arguments) != 1 {
		in.errorAt(e, "conversion to %s expects 1 argument, got %d", target.Name(), len(e.Arguments))
		in.inferArguments(e.Arguments, nil)
		return target
	}

	arg := e.Arguments[0]
	actual := in.inferExpression(arg, nil)
	switch {
	case types.IsUntypedInteger(actual):
		in.convertUntyped(arg, target)
	case types.IsInteger(actual), types.IsInvalid(actual):
	default:
		in.errorAt(e, "cannot convert %s to %s", actual.Name(), target.Name())
	}
	return target
}

func (in *Inferrer) inferMemberExpression(e *ast.MemberExpression) types.Type {
	member, found := in.lookupMember(e)
	if !found {
		return types.InvalidType
	}
	if member.Kind == types.MemberProperty && len(member.ReturnTypes) == 1 {
		return member.ReturnTypes[0]
	}
	return types.NewFunction(member.Parameters, member.ReturnTypes)
}

// lookupMember finds a member on the receiver's type, or a type function when
// the receiver names a struct.
func (in *Inferrer) lookupMember(e *ast.MemberExpression) (types.MemberDefinition, bool) {
	if ident, ok := e.Receiver.(*ast.Identifier); ok {
		if st, isStruct := structOf(ident); isStruct {
			setType(ident, st)
			return st.LookupTypeMember(e.Name.Value)
		}
	}

	receiver := in.inferDefault(e.Receiver)
	provider, ok := underlying(receiver).(types.MemberProvider)
	if !ok {
		return types.MemberDefinition{}, false
	}
	return provider.LookupMember(e.Name.Value)
}

// inferIndexExpression types collection access. Arrays and slices are
// indexed by int, following Go; maps by their key type.
func (in *Inferrer) inferIndexExpression(e *ast.IndexExpression) types.Type {
	switch collection := underlying(in.inferDefault(e.Left)).(type) {
	case *types.Array:
		in.inferValue(e.Index, types.IntType)
		return collection.ElementType()
	case *types.Slice:
		in.inferValue(e.Index, types.IntType)
		return collection.ElementType()
	case *types.Map:
		in.inferValue(e.Index, collection.Key)
		return collection.Value
	default:
		in.inferDefault(e.Index)
		return types.InvalidType
	}
}

// branch is the value an if branch, block or loop produces.
type branch struct {
	typ     types.Type
	expr    ast.Expression
	returns bool
}

// inferIfExpression infers an if. As a value, its non-returning branches
// must agree; as a statement, each branch is inferred on its own.
func (in *Inferrer) inferIfExpression(e *ast.IfExpression, expected types.Type, asValue bool) branch {
	in.inferDefault(e.Condition)

	branches := []branch{in.inferBlock(e.Body, expected, asValue)}
	switch elseBranch := e.Else.(type) {
	case *ast.BlockExpression:
		branches = append(branches, in.inferBlock(elseBranch, expected, asValue))
	case *ast.IfExpression:
		branches = append(branches, in.inferIfExpression(elseBranch, expected, asValue))
	}

	result := branch{typ: types.VoidType, expr: e, returns: len(branches) == 2}
	for _, b := range branches {
		result.returns = result.returns && b.returns
	}

	if asValue {
		values := []branch{}
		for _, b := range branches {
			if !b.returns {
				values = append(values, b)
			}
		}
		if expected != nil {
			// Each branch has already been checked against the expected type.
			result.typ = expected
		} else {
			result.typ = in.unify(values, "if branch")
		}
	}

	setType(e, result.typ)
	return result
}

// inferBlock infers a block's statements. As a value, its final expression
// statement is its value and is inferred with the expected type.
func (in *Inferrer) inferBlock(block *ast.BlockExpression, expected types.Type, asValue bool) branch {
	if block == nil {
		return branch{typ: types.VoidType}
	}

	result := branch{typ: types.VoidType}
	tail := blockTail(block)
	last := len(block.Statements) - 1
	for i, stmt := range block.Statements {
		if _, ok := stmt.(*ast.ReturnStatement); ok {
			result.returns = true
		}
		if asValue && tail != nil && i == last {
			result.expr = tail
			if expected != nil {
				result.typ = in.inferValue(tail, expected)
			} else {
				result.typ = in.inferExpression(tail, nil)
			}
			continue
		}
		in.inferStatement(stmt)
	}

	setType(block, result.typ)
	return result
}

// blockTail returns the expression that ends a block, if it ends in one.
func blockTail(block *ast.BlockExpression) ast.Expression {
	if block == nil || len(block.Statements) == 0 {
		return nil
	}
	if stmt, ok := block.Statements[len(block.Statements)-1].(*ast.ExpressionStatement); ok {
		return stmt.Expression
	}
	return nil
}

// unify finds the type several values share. Untyped constants take the type
// of the first typed value, which is an error unless it is an integer, and
// integers of different types are an error. Other mismatches are left for the
// analyzer.
func (in *Inferrer) unify(values []branch, what string) types.Type {
	var target types.Type
	for _, v := range values {
		if types.IsInteger(v.typ) {
			target = v.typ
			break
		}
	}
	if target == nil {
		for _, v := range values {
			// A branch without a value is reported by the analyzer.
			if v.typ != nil && !types.IsInvalid(v.typ) && !types.IsUntypedInteger(v.typ) && !types.IsVoid(v.typ) {
				target = v.typ
				break
			}
		}
	}

	if target == nil {
		if len(values) > 0 {
			return values[0].typ
		}
		return types.VoidType
	}

	for _, v := range values {
		switch {
		case types.IsUntypedInteger(v.typ):
			in.convertUntyped(v.expr, target)
		case types.IsInteger(v.typ) && !types.IsTypesEqual(v.typ, target):
			in.errorAt(v.expr, "mismatched %s types %s and %s", what, target.Name(), v.typ.Name())
		}
	}
	return target
}

// inferForLoop infers the controller and body. A loop's value is the value
// its breaks carry.
func (in *Inferrer) inferForLoop(loop *ast.ForLoop, expected types.Type) types.Type {
	switch c := loop.Controller.(type) {
	case *ast.While:
		in.inferDefault(c.Condition)
	case *ast.Constructed:
		in.inferStatement(c.Initializer)
		in.inferDefault(c.Condition)
		in.inferStatement(c.Iterator)
	case *ast.Range:
		in.inferRange(c)
	}

	ctx := &loopContext{loop: loop, expected: expected}
	in.loops = append(in.loops, ctx)
	if loop.Body != nil {
		in.inferStatements(loop.Body.Statements)
	}
	in.loops = in.loops[:len(in.loops)-1]

	in.loopBreaks[loop] = ctx.values
	if len(ctx.values) == 0 {
		return types.VoidType
	}
	if expected != nil {
		// Each break value has already been checked against the expected type.
		return expected
	}

	values := make([]branch, len(ctx.values))
	for i, value := range ctx.values {
		values[i] = branch{typ: typeOf(value), expr: value}
	}
	return in.unify(values, "break value")
}

// inferRange types range variables. Arrays and slices produce an int index,
// following Go. A single variable holds the (key, value) tuple.
func (in *Inferrer) inferRange(r *ast.Range) {
	key, value := types.Type(types.InvalidType), types.Type(types.InvalidType)
	switch collection := underlying(in.inferDefault(r.Iterable)).(type) {
	case *types.Array:
		key, value = types.IntType, collection.ElementType()
	case *types.Slice:
		key, value = types.IntType, collection.ElementType()
	case *types.Map:
		key, value = collection.Key, collection.Value
	}

	switch len(r.Variables) {
	case 1:
		r.Variables[0].SetResolvedType(&types.Tuple{Types: []types.Type{key, value}})
	case 2:
		r.Variables[0].SetResolvedType(key)
		r.Variables[1].SetResolvedType(value)
	default:
		for _, variable := range r.Variables {
			variable.SetResolvedType(types.InvalidType)
		}
	}
}

func (in *Inferrer) inferFunctionLiteral(fnLit *ast.FunctionLiteral) types.Type {
	signature := in.functionSignature(fnLit.Parameters, fnLit.ReturnTypes)
	setType(fnLit, signature)
	in.inferFunctionBody(fnLit.Parameters, signature, fnLit.Body)
	return signature
}

func (in *Inferrer) inferElements(elements []ast.Expression, elementType types.Type) {
	for _, element := range elements {
		in.inferValue(element, elementType)
	}
}

func (in *Inferrer) inferMapLiteral(e *ast.MapLiteral) types.Type {
	in.resolveNamed(e.Type)
	mapType, ok := e.Type.(*types.Map)
	if !ok {
		for _, pair := range e.Pairs {
			in.inferDefault(pair.Key)
			in.inferDefault(pair.Value)
		}
		return types.InvalidType
	}

	for _, pair := range e.Pairs {
		in.inferValue(pair.Key, mapType.Key)
		in.inferValue(pair.Value, mapType.Value)
	}
	return mapType
}

func (in *Inferrer) inferStructLiteral(e *ast.StructLiteral) types.Type {
	st, ok := structOf(e.Name)
	for _, field := range e.Fields {
		var fieldType types.Type
		if ok {
			if definition, exists := st.Fields[field.Name.Value]; exists {
				fieldType = definition.Type
			}
		}
		if fieldType != nil {
			in.inferValue(field.Value, fieldType)
		} else {
			in.inferDefault(field.Value)
		}
	}

	if !ok {
		return types.InvalidType
	}
	setType(e.Name, st)
	return st
}

func isBuiltin(ident *ast.Identifier) bool {
	sym, ok := ident.Symbol.(*resolver.Symbol)
	return ok && sym.Kind == resolver.BuiltinSymbol
}
