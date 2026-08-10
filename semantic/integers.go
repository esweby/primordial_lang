package semantic

import (
	"fmt"
	"math/big"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func integerConstantValue(expression ast.Expression) (*big.Int, bool) {
	switch expression := expression.(type) {
	case *ast.IntegerLiteral:
		return new(big.Int).Set(expression.Value), true
	case *ast.PrefixExpression:
		value, ok := integerConstantValue(expression.Right)
		if !ok {
			return nil, false
		}
		switch expression.Operator {
		case "-":
			return new(big.Int).Neg(value), true
		default:
			return nil, false
		}
	case *ast.InfixExpression:
		left, leftOK := integerConstantValue(expression.Left)
		right, rightOK := integerConstantValue(expression.Right)
		if !leftOK || !rightOK {
			return nil, false
		}
		switch expression.Operator {
		case "+":
			return new(big.Int).Add(left, right), true
		case "-":
			return new(big.Int).Sub(left, right), true
		case "*":
			return new(big.Int).Mul(left, right), true
		case "/":
			if right.Sign() == 0 {
				return nil, false
			}
			return new(big.Int).Quo(left, right), true
		default:
			return nil, false
		}
	default:
		return nil, false
	}
}

func coerceIntegerConstant(expression ast.Expression, target types.Type) error {
	integerType, ok := target.(*types.Integer)
	if !ok {
		return fmt.Errorf("cannot use untyped integer as %s", target.Name())
	}
	if ifExpression, ok := expression.(*ast.IfExpression); ok {
		return coerceIfIntegerBranches(ifExpression, target)
	}

	value, ok := integerConstantValue(expression)
	if !ok {
		return fmt.Errorf("integer expression is not a compile-time constant")
	}
	if !integerType.CanRepresent(value) {
		return fmt.Errorf("integer constant %s is not representable as %s", value.String(), target.Name())
	}
	if typed, ok := expression.(ast.TypedExpression); ok {
		typed.SetResolvedType(target)
	}
	return nil
}

func coerceIfIntegerBranches(expression *ast.IfExpression, target types.Type) error {
	if err := coerceBlockIntegerValue(expression.Body, target); err != nil {
		return err
	}
	switch elseBranch := expression.Else.(type) {
	case *ast.BlockExpression:
		return coerceBlockIntegerValue(elseBranch, target)
	case *ast.IfExpression:
		return coerceIfIntegerBranches(elseBranch, target)
	default:
		return fmt.Errorf("if expression has no integer-producing else branch")
	}
}

func coerceBlockIntegerValue(block *ast.BlockExpression, target types.Type) error {
	if block == nil || len(block.Statements) == 0 {
		return fmt.Errorf("integer-producing block is empty")
	}
	lastStatement := block.Statements[len(block.Statements)-1]
	if _, returns := lastStatement.(*ast.ReturnStatement); returns {
		return nil
	}
	statement, ok := lastStatement.(*ast.ExpressionStatement)
	if !ok {
		return fmt.Errorf("integer-producing block must end with an expression")
	}
	return coerceIntegerConstant(statement.Expression, target)
}

func (sa *SemanticAnalyzer) requireAssignable(target, actual types.Type, expression ast.Expression) error {
	if types.IsUntypedInteger(actual) {
		return coerceIntegerConstant(expression, target)
	}
	if !types.IsTypesEqual(target, actual) {
		return fmt.Errorf("expected %s, got %s", target.Name(), actual.Name())
	}
	return nil
}

func (sa *SemanticAnalyzer) defaultInteger(expression ast.Expression, actual types.Type) (types.Type, error) {
	if !types.IsUntypedInteger(actual) {
		return actual, nil
	}
	if err := coerceIntegerConstant(expression, types.Int64Type); err != nil {
		return types.InvalidType, err
	}
	return types.Int64Type, nil
}

func (sa *SemanticAnalyzer) resolveIntegerOperands(
	expression *ast.InfixExpression,
	leftType types.Type,
	rightType types.Type,
) (types.Type, error) {
	leftUntyped := types.IsUntypedInteger(leftType)
	rightUntyped := types.IsUntypedInteger(rightType)

	switch {
	case leftUntyped && rightUntyped:
		expression.SetResolvedType(types.UntypedIntegerType)
		return types.UntypedIntegerType, nil
	case leftUntyped && types.IsInteger(rightType):
		if err := coerceIntegerConstant(expression.Left, rightType); err != nil {
			return types.InvalidType, err
		}
		expression.SetResolvedType(rightType)
		return rightType, nil
	case types.IsInteger(leftType) && rightUntyped:
		if err := coerceIntegerConstant(expression.Right, leftType); err != nil {
			return types.InvalidType, err
		}
		expression.SetResolvedType(leftType)
		return leftType, nil
	case types.IsInteger(leftType) && types.IsInteger(rightType):
		if !types.IsTypesEqual(leftType, rightType) {
			return types.InvalidType, fmt.Errorf("mismatched integer types: %s and %s", leftType.Name(), rightType.Name())
		}
		expression.SetResolvedType(leftType)
		return leftType, nil
	default:
		return types.InvalidType, fmt.Errorf("mismatched types: %s and %s", leftType.Name(), rightType.Name())
	}
}
