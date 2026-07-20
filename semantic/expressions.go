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
		if _, ok := stmt.(*ast.ExpressionStatement); ok {
			lastType = stmtType
		}
		// Other statements (declarations, assignments, function defs) do not affect the block's value.
	}

	if returns {
		return BlockResult{Type: nil, Returns: true}
	}
	return BlockResult{Type: lastType, Returns: false}
}

func (sa *SemanticAnalyzer) analyzeExpression(exp ast.Expression) types.Type {
	if exp == nil {
		sa.error("analyzeExpression called with nil expression")
		return types.InvalidType
	}

	switch e := exp.(type) {
	case *ast.IntegerLiteral:
		return types.Int64Type
	case *ast.Boolean:
		return types.BoolType
	// case *ast.StringLiteral:
	// return types.StringType
	case *ast.Identifier:
		sym, ok := sa.current.Get(e.Value)
		if !ok {
			sa.error("undefined identifier: " + e.Value)
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
		if sa.isInvalidInfixType(left) {
			sa.error(fmt.Sprintf("invalid type: %s", left.Name()))
			return types.InvalidType
		}
		if sa.isInvalidInfixType(right) {
			sa.error(fmt.Sprintf("invalid type: %s", right.Name()))
			return types.InvalidType
		}
		if !types.IsNumeric(left) || !types.IsNumeric(right) {
			sa.error(fmt.Sprintf("mismatched types: %s and %s", left.Name(), right.Name()))
			return types.InvalidType
		}
		return left
	case "<=", "<", ">", ">=":
		if !types.IsNumeric(left) {
			sa.error(fmt.Sprintf("invalid type: %s", left.Name()))
			return types.InvalidType
		}
		if !types.IsNumeric(right) {
			sa.error(fmt.Sprintf("invalid type: %s", right.Name()))
			return types.InvalidType
		}

		if !types.IsTypesEqual(left, right) {
			sa.error(fmt.Sprintf("mismatched types: %s %s %s", left.Name(), e.Operator, right.Name()))
			return types.InvalidType
		}

		return types.BoolType
	case "==", "!=":
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
		if !types.IsNumeric(right) {
			sa.error(fmt.Sprintf("analyzePrefixExpression - operator: type is not numeric. Got=%s", right.Name()))
			return types.InvalidType
		}
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
	// First, analyze the callee expression.
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
		if !types.IsAssignable(p.Type, argType) {
			sa.error(fmt.Sprintf("argument %d: expected %s, got %s",
				i, p.Type.Name(), argType.Name()))
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

func (sa *SemanticAnalyzer) analyzeIfExpression(ifExpr *ast.IfExpression, expectsValue bool) types.Type {
	// 1. Analyze condition.
	condType := sa.analyzeExpression(ifExpr.Condition)
	if !types.IsBoolean(condType) {
		sa.error("if condition must be boolean")
		return types.InvalidType
	}

	// 2. Analyze 'then' block.
	thenResult := sa.analyzeBlock(ifExpr.Body)

	// 3. Analyze 'else' branch if present.
	var elseResult BlockResult
	hasElse := ifExpr.Else != nil
	if hasElse {
		switch elseBranch := ifExpr.Else.(type) {
		case *ast.BlockExpression:
			elseResult = sa.analyzeBlock(elseBranch)
		case *ast.IfExpression:
			// Recursively analyze else‑if; passes expectsValue down.
			elseType := sa.analyzeIfExpression(elseBranch, expectsValue)
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

func (sa *SemanticAnalyzer) isInvalidInfixType(t types.Type) bool {
	return types.IsBoolean(t) || types.IsFunction(t) || types.IsString(t)
}

func (sa *SemanticAnalyzer) isInvalidPrefixType(t types.Type) bool {
	return !types.IsBoolean(t) && !types.IsNumeric(t)
}
