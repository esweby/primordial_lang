package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (sa *SemanticAnalyzer) analyzeBlock(block *ast.BlockExpression) BlockResult {
	if block == nil {
		return BlockResult{Type: nil, Returns: false}
	}

	var lastType types.Type
	var lastExpression ast.Expression
	var returns bool

	for _, stmt := range block.Statements {
		stmtType := sa.analyzeStatement(stmt)

		// If the statement is a return statement, mark the block as returning.
		if _, ok := stmt.(*ast.ReturnStatement); ok {
			returns = true
			// Do not update lastType on return.
			continue
		}

		// If the statement is an expression statement, its type may be the block's value.
		if expressionStatement, ok := stmt.(*ast.ExpressionStatement); ok {
			lastType = stmtType
			lastExpression = expressionStatement.Expression
		}
		// Other statements (declarations, assignments, function defs) do not affect the block's value.
	}

	if returns {
		return BlockResult{Type: nil, Returns: true}
	}
	return BlockResult{Type: lastType, Returns: false, Expression: lastExpression}
}

func (sa *SemanticAnalyzer) analyzeExpression(exp ast.Expression) types.Type {
	if exp == nil {
		sa.error("analyzeExpression called with nil expression")
		return types.InvalidType
	}

	switch e := exp.(type) {
	case *ast.IntegerLiteral:
		if e.GetResolvedType() != nil {
			return e.GetResolvedType()
		}
		return types.UntypedIntegerType
	case *ast.Boolean:
		return types.BoolType
	case *ast.StringLiteral:
		return types.StringType
	case *ast.Identifier:
		sym, ok := sa.current.Get(e.Value)
		if !ok {
			sa.error("undefined identifier: " + e.Value)
			return types.InvalidType
		}
		if _, isType := sym.(*StructSymbol); isType {
			sa.error(fmt.Sprintf("type %s cannot be used as a value", e.Value))
			return types.InvalidType
		}
		return sym.Type()
	case *ast.InfixExpression:
		return sa.analyzeInfixExpression(e)
	case *ast.PrefixExpression:
		return sa.analyzePrefixExpression(e)
	case *ast.CallExpression:
		return sa.analyzeCallExpression(e)
	case *ast.IfExpression:
		// When used as an expression, we expect a value.
		return sa.analyzeIfExpression(e, true)
	case *ast.FunctionLiteral:
		// A function literal as a standalone expression (e.g., passed as argument).
		// We analyze it but do not register it; we return its generic type.
		sa.analyzeStandaloneFunctionLiteral(e)
		return types.FunctionType
	case *ast.ArrayLiteral:
		return sa.analyzeArrayLiteral(e)
	case *ast.SliceLiteral:
		return sa.analyzeSliceLiteral(e)
	case *ast.MemberExpression:
		return sa.analyzeMemberExpression(e)
	case *ast.IndexExpression:
		return sa.analyzeIndexExpression(e)
	case *ast.StructLiteral:
		return sa.analyzeStructLiteral(e)
	case *ast.MapLiteral:
		return sa.analyzeMapLiteral(e)
	default:
		sa.error(fmt.Sprintf("analyzeExpression received unexpected expression: %T", e))
		return types.InvalidType
	}
}

func (sa *SemanticAnalyzer) analyzeInfixExpression(e *ast.InfixExpression) types.Type {
	left := sa.analyzeExpression(e.Left)
	right := sa.analyzeExpression(e.Right)

	if types.IsInvalid(left) || types.IsInvalid(right) {
		return types.InvalidType
	}

	switch e.Operator {
	case "+", "-", "*", "/":
		if e.Operator == "+" && types.IsString(left) && types.IsString(right) {
			return types.StringType
		}
		if (types.IsInteger(left) || types.IsUntypedInteger(left)) &&
			(types.IsInteger(right) || types.IsUntypedInteger(right)) {
			resultType, err := sa.resolveIntegerOperands(e, left, right)
			if err != nil {
				sa.error(err.Error())
				return types.InvalidType
			}
			if e.Operator == "/" {
				if divisor, ok := integerConstantValue(e.Right); ok && divisor.Sign() == 0 {
					sa.error("division by zero")
					return types.InvalidType
				}
			}
			return resultType
		}
		if !types.IsNumeric(left) || !types.IsNumeric(right) || !types.IsTypesEqual(left, right) {
			sa.error(fmt.Sprintf("mismatched types: %s and %s", left.Name(), right.Name()))
			return types.InvalidType
		}
		e.SetResolvedType(left)
		return left
	case "<=", "<", ">", ">=":
		if (types.IsInteger(left) || types.IsUntypedInteger(left)) &&
			(types.IsInteger(right) || types.IsUntypedInteger(right)) {
			if _, err := sa.resolveIntegerOperands(e, left, right); err != nil {
				sa.error(err.Error())
				return types.InvalidType
			}
			return types.BoolType
		}
		if !types.IsNumeric(left) || !types.IsNumeric(right) || !types.IsTypesEqual(left, right) {
			sa.error(fmt.Sprintf("mismatched types: %s %s %s", left.Name(), e.Operator, right.Name()))
			return types.InvalidType
		}

		return types.BoolType
	case "==", "!=":
		if (types.IsInteger(left) || types.IsUntypedInteger(left)) &&
			(types.IsInteger(right) || types.IsUntypedInteger(right)) {
			if _, err := sa.resolveIntegerOperands(e, left, right); err != nil {
				sa.error(err.Error())
				return types.InvalidType
			}
			return types.BoolType
		}
		comparable := types.IsNumeric(left) ||
			types.IsBoolean(left) ||
			types.IsString(left)

		if !comparable {
			sa.error(fmt.Sprintf("invalid type: %s", left.Name()))
			return types.InvalidType
		}

		if !types.IsTypesEqual(left, right) {
			sa.error(fmt.Sprintf("mismatched types: %s %s %s", left.Name(), e.Operator, right.Name()))
			return types.InvalidType
		}

		return types.BoolType
	default:
		sa.error(fmt.Sprintf("unknown infix operator: %s", e.Operator))
		return types.InvalidType
	}
}

func (sa *SemanticAnalyzer) analyzePrefixExpression(pe *ast.PrefixExpression) types.Type {
	right := sa.analyzeExpression(pe.Right)

	if sa.isInvalidPrefixType(right) {
		sa.error(fmt.Sprintf("analyzePrefixExpression: type is not a valid Prefix Expression: %s", right.Name()))
		return types.InvalidType
	}

	switch pe.Operator {
	case "-":
		if types.IsUntypedInteger(right) {
			pe.SetResolvedType(types.UntypedIntegerType)
			return types.UntypedIntegerType
		}
		if !types.IsNumeric(right) {
			sa.error(fmt.Sprintf("analyzePrefixExpression - operator: type is not numeric. Got=%s", right.Name()))
			return types.InvalidType
		}
		if integerType, ok := right.(*types.Integer); ok && !integerType.Signed() {
			sa.error(fmt.Sprintf("cannot negate unsigned integer %s", right.Name()))
			return types.InvalidType
		}
		pe.SetResolvedType(right)
		return right
	case "!":
		if !types.IsBoolean(right) {
			sa.error(fmt.Sprintf("analyzePrefixExpression ! operator: type is not boolean. Got=%s", right.Name()))
			return types.InvalidType
		}
		return right
	default:
		sa.error(fmt.Sprintf("unknown prefix operator: %s", pe.Operator))
		return types.InvalidType
	}
}

func (sa *SemanticAnalyzer) analyzeCallExpression(ce *ast.CallExpression) types.Type {
	if member, ok := ce.Function.(*ast.MemberExpression); ok {
		return sa.analyzeMemberCall(member, ce.Arguments)
	}

	if ident, ok := ce.Function.(*ast.Identifier); ok {
		if target, isBuiltinType := types.GetBuiltin(ident.Value); isBuiltinType && types.IsInteger(target) {
			return sa.analyzeIntegerConversion(ce, target)
		}
		if _, declared := sa.current.Get(ident.Value); !declared && ident.Value == "len" {
			return sa.analyzeLenCall(ce.Arguments)
		}
	}

	calleeType := sa.analyzeExpression(ce.Function)
	if !types.IsFunction(calleeType) {
		sa.error("cannot call non-function")
		return types.InvalidType
	}

	// For now, we only support calls where the callee is an identifier.
	// This is sufficient for your current AST/parser.
	ident, ok := ce.Function.(*ast.Identifier)
	if !ok {
		sa.error("call expression: callee is not an identifier (unsupported)")
		return types.InvalidType
	}

	// Look up the function symbol.
	sym, ok := sa.current.Get(ident.Value)
	if !ok {
		sa.error(fmt.Sprintf("undefined function: %s", ident.Value))
		return types.InvalidType
	}

	fs, isNamedFunction := sym.(*FunctionSymbol)
	if !isNamedFunction {
		if !types.IsFunction(sym.Type()) {
			sa.error(fmt.Sprintf("symbol '%s' is not a function", ident.Value))
			return types.InvalidType
		}

		// A `function` value is callable, but its signature is currently unknown.
		for _, arg := range ce.Arguments {
			sa.analyzeExpression(arg)
		}

		return types.FunctionType
	}

	// Check argument count.
	if len(ce.Arguments) != len(fs.params) {
		sa.error(fmt.Sprintf("wrong number of arguments: expected %d, got %d",
			len(fs.params), len(ce.Arguments)))
		return types.InvalidType
	}

	// Check argument types.
	for i, p := range fs.params {
		argType := sa.analyzeExpression(ce.Arguments[i])
		if err := sa.requireAssignable(p.Type, argType, ce.Arguments[i]); err != nil {
			sa.error(fmt.Sprintf("argument %d: %s", i, err.Error()))
			return types.InvalidType
		}
	}

	// Return the call's result type.
	switch len(fs.returnTypes) {
	case 0:
		return types.VoidType
	case 1:
		return fs.returnTypes[0]
	default:
		return &types.Tuple{
			Types: fs.returnTypes,
		}
	}
}

func (sa *SemanticAnalyzer) analyzeIntegerConversion(call *ast.CallExpression, target types.Type) types.Type {
	if len(call.Arguments) != 1 {
		sa.error(fmt.Sprintf("integer conversion to %s expects 1 argument, got %d", target.Name(), len(call.Arguments)))
		return types.InvalidType
	}
	actual := sa.analyzeExpression(call.Arguments[0])
	if types.IsUntypedInteger(actual) {
		if err := coerceIntegerConstant(call.Arguments[0], target); err != nil {
			sa.error(err.Error())
			return types.InvalidType
		}
	} else if !types.IsInteger(actual) {
		sa.error(fmt.Sprintf("cannot convert %s to %s", actual.Name(), target.Name()))
		return types.InvalidType
	}
	call.SetResolvedType(target)
	return target
}

func (sa *SemanticAnalyzer) analyzeIfExpression(ifExpr *ast.IfExpression, expectsValue bool) types.Type {
	// 1. Analyze condition.
	condType := sa.analyzeExpression(ifExpr.Condition)
	if !types.IsBoolean(condType) {
		sa.error("if condition must be boolean")
		return types.InvalidType
	}

	// 2. Analyze 'then' block.
	thenResult := sa.analyzeScopedBlock(ifExpr.Body)

	// 3. Analyze 'else' branch if present.
	var elseResult BlockResult
	hasElse := ifExpr.Else != nil
	if hasElse {
		switch elseBranch := ifExpr.Else.(type) {
		case *ast.BlockExpression:
			elseResult = sa.analyzeScopedBlock(elseBranch)
		case *ast.IfExpression:
			// Recursively analyze else‑if; passes expectsValue down.
			sa.enterScope()
			elseType := sa.analyzeIfExpression(elseBranch, expectsValue)
			sa.exitScope()
			elseResult = BlockResult{Type: elseType, Returns: false} // approximate, may need refinement
		default:
			sa.error("unexpected else structure")
			return types.InvalidType
		}
	}

	// 4. Context-specific checks.
	if expectsValue {
		if !hasElse {
			sa.error("if expression used as value requires an else branch")
			return types.InvalidType
		}

		var overallType types.Type

		switch {
		case thenResult.Returns && elseResult.Returns:
			// Both branches return → no value produced.
			sa.error("if expression branches both return, no value produced")
			return types.InvalidType

		case thenResult.Returns:
			// then returns, else must yield a value.
			if elseResult.Type == nil {
				sa.error("else branch must yield an expression when then branch returns")
				return types.InvalidType
			}
			overallType = elseResult.Type

		case elseResult.Returns:
			// else returns, then must yield a value.
			if thenResult.Type == nil {
				sa.error("then branch must yield an expression when else branch returns")
				return types.InvalidType
			}
			overallType = thenResult.Type

		default:
			// Neither returns: both must end with an expression and match types.
			if thenResult.Type == nil || elseResult.Type == nil {
				sa.error("if expression branches must end with an expression")
				return types.InvalidType
			}
			if types.IsUntypedInteger(thenResult.Type) && types.IsInteger(elseResult.Type) {
				if err := coerceIntegerConstant(thenResult.Expression, elseResult.Type); err != nil {
					sa.error(err.Error())
					return types.InvalidType
				}
				thenResult.Type = elseResult.Type
			} else if types.IsInteger(thenResult.Type) && types.IsUntypedInteger(elseResult.Type) {
				if err := coerceIntegerConstant(elseResult.Expression, thenResult.Type); err != nil {
					sa.error(err.Error())
					return types.InvalidType
				}
				elseResult.Type = thenResult.Type
			}
			if !types.IsTypesEqual(thenResult.Type, elseResult.Type) {
				sa.error(fmt.Sprintf("if expression branches have mismatched types: %s vs %s",
					thenResult.Type.Name(), elseResult.Type.Name()))
				return types.InvalidType
			}
			overallType = thenResult.Type
		}
		return overallType
	}
	// Statement context: we don't care about the type, but we already validated.
	return nil
}

func (sa *SemanticAnalyzer) analyzeIndexExpression(exp *ast.IndexExpression) types.Type {
	collectionType := sa.analyzeExpression(exp.Left)
	indexType := sa.analyzeExpression(exp.Index)
	if types.IsUntypedInteger(indexType) {
		resolved, err := sa.defaultInteger(exp.Index, indexType)
		if err != nil {
			sa.error(err.Error())
			return types.InvalidType
		}
		indexType = resolved
	}
	if !types.IsInteger(indexType) {
		sa.error(fmt.Sprintf("collection index must be an integer, got %s", indexType.Name()))
		return types.InvalidType
	}

	switch collection := collectionType.(type) {
	case *types.Array:
		return collection.ElementType()
	case *types.Slice:
		return collection.ElementType()
	default:
		sa.error(fmt.Sprintf("type %s cannot be indexed", collectionType.Name()))
		return types.InvalidType
	}
}

func (sa *SemanticAnalyzer) analyzeLenCall(args []ast.Expression) types.Type {
	if len(args) != 1 {
		sa.error(fmt.Sprintf("len expects 1 argument, got %d", len(args)))
		return types.InvalidType
	}

	argumentType := sa.analyzeExpression(args[0])
	switch argumentType.Kind() {
	case types.KindArray, types.KindSlice, types.KindString:
		return types.Int64Type
	default:
		sa.error(fmt.Sprintf("argument to len not supported, got %s", argumentType.Name()))
		return types.InvalidType
	}
}

func (sa *SemanticAnalyzer) isInvalidInfixType(t types.Type) bool {
	return types.IsBoolean(t) || types.IsFunction(t) || types.IsString(t)
}

func (sa *SemanticAnalyzer) isInvalidPrefixType(t types.Type) bool {
	return !types.IsBoolean(t) && !types.IsNumeric(t)
}
