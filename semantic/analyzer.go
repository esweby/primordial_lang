package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

type Analyzer struct {
	program *ast.Program
	errors []error
	current *SymbolTable
}

func New(program *ast.Program) Analyzer {
	return Analyzer{
		program: program,
		errors: []error{},
		current: NewSymbolTable(),
	}
}

func (a *Analyzer) Analyze() []error {
	for _, stmt := range a.program.Statements {
		a.checkStatement(stmt)
	}

	return a.errors
}

func (a *Analyzer) checkStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.DeclareStatement: {
		a.analzyzeDeclareStatement(s)
	}
	case *ast.ReturnStatement: {
		expr := s.ReturnValue
		a.analyzeExpression(expr)
	}
	case *ast.ExpressionStatement: {
		expr := s.Expression
		a.analyzeExpression(expr)
	}
	case *ast.FunctionStatement: {}
	default:
		a.error(fmt.Sprintf("evaluateStatement got unexpected statement: %T", stmt))
	}
}

func (a *Analyzer) analyzeExpression(exp ast.Expression) types.Type {
	switch e := exp.(type) {
	case *ast.IntegerLiteral:
		return types.Int64Type
	case *ast.Boolean:
		return types.BoolType
	case *ast.Identifier:
		return types.InvalidType
	// Evaluating Infix Expressions
	// Reminder: An InfixExpression is 
	case *ast.InfixExpression:
		return a.analyzeInfixExpression(e)
	case *ast.CallExpression:
		return a.analyzeCallExpression(e)
	}

	a.error(fmt.Sprintf("analyzeExpression: unknown expression type: %T", exp))
	return types.InvalidType
}

func (a *Analyzer) analzyzeDeclareStatement(s *ast.DeclareStatement) {
	isTypeDefined := s.Type != nil 

	right := a.analyzeExpression(s.Value)
	if isTypeDefined {
		left := s.Type
		if !types.IsTypesEqual(left, right) {
			a.error(
				fmt.Sprintf(
					"analzyzeDeclareStatement: invalid type declaration: %s + %s", 
					left.Name(), right.Name(),
				),
			)
		}
	} else {
		s.Type = right
	}
}

func (a *Analyzer) analyzeInfixExpression(e *ast.InfixExpression) types.Type {
	left := a.analyzeExpression(e.Left)
	right := a.analyzeExpression(e.Right)

	switch e.Operator {
	case "+", "-", "*", "/", "<", ">":
		if types.IsBoolean(left) || types.IsFunction(left) || types.IsString(left) {
			a.error(fmt.Sprintf("analyzeInfixExpression: left infixExpression is invalid type: %s", left.Name()))
			return types.InvalidType
		}

		if types.IsBoolean(right) || types.IsFunction(right) || types.IsString(right) {
			a.error(fmt.Sprintf("analyzeInfixExpression: right infixExpression is invalid type: %s", right.Name()))
			return types.InvalidType
		}

		if !types.IsTypesEqual(left, right) {
			a.error(fmt.Sprintf("analyzeInfixExpression: found invalid type in infixExpression: %s + %s", left.Name(), right.Name()))
			return types.InvalidType
		}
	
		return left
	case "==", "!=":
		if types.IsFunction(left) || types.IsString(left) {
			a.error(fmt.Sprintf("analyzeInfixExpression: left infixExpression is invalid type: %s", left.Name()))
			return types.InvalidType
		}

		if types.IsFunction(right) || types.IsString(right) {
			a.error(fmt.Sprintf("analyzeInfixExpression: right infixExpression is invalid type: %s", right.Name()))
			return types.InvalidType
		}

		if !types.IsTypesEqual(left, right) {
			a.error(fmt.Sprintf("analyzeInfixExpression: found invalid type in infixExpression: %s + %s", left.Name(), right.Name()))
			return types.InvalidType
		}
	
		return left
	default:
		a.error(fmt.Sprintf("analyzeInfixExpression: fallen through to default: %s", e.Operator))
		return types.InvalidType
	}

}

func (a *Analyzer) analyzeCallExpression(e *ast.CallExpression) types.Type {
	// lookup symbol table 

	// check the functions registered

	// check the arguments

	return types.FunctionType
}

// func (a *Analyzer) inferType(value *ast.Expression) types.Type {}

func (a *Analyzer) enterScope() {
	a.current = NewEnclosedSymbolTable(a.current)
}

func (a *Analyzer) exitScope() {
	a.current = a.current.outer
}

func (a *Analyzer) error(err string) {
	a.errors = append(a.errors, fmt.Errorf("%s", err))
}