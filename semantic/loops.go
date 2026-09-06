package semantic

import (
	"fmt"
	"slices"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (sa *SemanticAnalyzer) analyzeForExpression(forExpr *ast.ForLoop, expectsValue bool) types.Type {
	sa.enterScope()
	defer sa.exitScope()

	if !sa.analyzeValidForController(forExpr) {
		return types.InvalidType
	}

	var breakValueType types.Type
	hasBreakValue := false
	hasBreakWithoutValue := false
	hasAnyBreak := false

	sa.loopDepth++
	if forExpr.Label != nil {
		sa.loopLabels = append(sa.loopLabels, forExpr.Label.Value)
	}

	defer func() {
		sa.loopDepth--
		if forExpr.Label != nil {
			sa.loopLabels = sa.loopLabels[:len(sa.loopLabels)-1]
		}
	}()

	for _, stmt := range forExpr.Body.Statements {
		if bs, ok := stmt.(*ast.BreakStatement); ok {
			hasAnyBreak = true
			if bs.Value != nil {
				vt := sa.analyzeExpression(bs.Value)
				if types.IsInvalid(vt) {
					return types.InvalidType
				}
				if types.IsUntypedInteger(vt) {
					resolved, err := sa.defaultInteger(bs.Value, vt)
					if err != nil {
						sa.error(err.Error())
						return types.InvalidType
					}
					vt = resolved
				}
				if hasBreakValue && !types.IsAssignable(breakValueType, vt) {
					sa.error("break value types must match")
					return types.InvalidType
				}
				breakValueType = vt
				hasBreakValue = true
			} else {
				hasBreakWithoutValue = true
			}
		}
		sa.analyzeStatement(stmt)
	}

	if expectsValue {
		if !hasAnyBreak {
			sa.error("loop used as expression must have a break to exit")
			return types.InvalidType
		}
		if hasBreakValue && hasBreakWithoutValue {
			sa.error("mixed break statements: some with values and some without")
			return types.InvalidType
		}
		if hasBreakValue {
			return breakValueType
		}
		return types.VoidType
	}

	if hasBreakValue && hasBreakWithoutValue {
		sa.error("mixed break statements in a statement loop")
		return types.InvalidType
	}
	return types.VoidType
}

func (sa *SemanticAnalyzer) analyzeValidForController(forExpr *ast.ForLoop) bool {
	switch c := forExpr.Controller.(type) {
	case *ast.While:
		return sa.analyzeWhileCondition(c)
	case *ast.Constructed:
		return sa.analyzeConstructedCondition(c)
	case *ast.Range:
		return sa.analyzeRangeCondition(forExpr)
	case *ast.Infinite:
		return true
	default:
		return false
	}
}

func (sa *SemanticAnalyzer) analyzeWhileCondition(cond *ast.While) bool {
	condType := sa.analyzeExpression(cond.Condition)
	if !types.IsBoolean(condType) {
		sa.error("while condition must be boolean")
		return false
	}

	return true
}

func (sa *SemanticAnalyzer) analyzeConstructedCondition(cond *ast.Constructed) bool {
	initType := sa.analyzeStatement(cond.Initializer)
	if types.IsInvalid(initType) {
		sa.error("invalid type in Initializer")
		return false
	}

	expr := sa.analyzeExpression(cond.Condition)
	if !types.IsBoolean(expr) {
		sa.error("Constructed condition must be boolean")
		return false
	}

	iter := sa.analyzeStatement(cond.Iterator)
	if types.IsInvalid(iter) {
		sa.error("invalid type as Iterator")
		return false
	}

	return true
}

func (sa *SemanticAnalyzer) analyzeRangeCondition(forExpr *ast.ForLoop) bool {
	ctrl := forExpr.Controller.(*ast.Range)

	iterType := sa.analyzeExpression(ctrl.Iterable)
	if types.IsInvalid(iterType) {
		return false
	}

	var firstType, valueType types.Type
	switch t := iterType.(type) {
	case *types.Array:
		firstType = types.Int64Type
		valueType = t.ElementType()
	case *types.Slice:
		firstType = types.Int64Type
		valueType = t.ElementType()
	case *types.Map:
		firstType = t.Key
		valueType = t.Value
	default:
		sa.error("range can only iterate over an array, slice, or map")
		return false
	}

	numVars := len(ctrl.Variables)
	if numVars < 1 || numVars > 2 {
		sa.error("range loop only allows for 1 or 2 variables")
		return false
	}

	switch numVars {
	case 2:
		sa.current.Set(ctrl.Variables[0].Value, &BasicSymbol{name: ctrl.Variables[0].Value, typ: firstType})
		sa.current.Set(ctrl.Variables[1].Value, &BasicSymbol{name: ctrl.Variables[1].Value, typ: valueType})
	case 1:
		tupleType := &types.Tuple{
			Types: []types.Type{firstType, valueType},
		}

		sa.current.Set(
			ctrl.Variables[0].Value,
			&BasicSymbol{
				name: ctrl.Variables[0].Value,
				typ:  tupleType,
			},
		)
	default:
		sa.error("range loop requires 1 or 2 variables")
		return false
	}
	return true
}

func (sa *SemanticAnalyzer) analyzeBreakStatement(brk *ast.BreakStatement) types.Type {
	if sa.loopDepth == 0 {
		sa.error("break outside of loop")
		return types.InvalidType
	}

	if brk.Label != nil {
		found := slices.Contains(sa.loopLabels, brk.Label.Value)
		if !found {
			sa.error(fmt.Sprintf("break label '%s' does not match any enclosing loop", brk.Label.Value))
			return types.InvalidType
		}
	}

	if brk.Value != nil {
		valueType := sa.analyzeExpression(brk.Value)
		if types.IsInvalid(valueType) {
			return types.InvalidType
		}
		// Default untyped integer to a concrete type
		if types.IsUntypedInteger(valueType) {
			// Use defaultInteger to coerce the expression's type
			resolved, err := sa.defaultInteger(brk.Value, valueType)
			if err != nil {
				sa.error(err.Error())
				return types.InvalidType
			}
			valueType = resolved
		}
		return valueType
	}

	return types.VoidType
}

func (sa *SemanticAnalyzer) analyzeContinueStatement(_ *ast.ContinueStatement) types.Type {
	if sa.loopDepth == 0 {
		sa.error("continue outside of loop")
		return types.InvalidType
	}

	return types.VoidType
}
