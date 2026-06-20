package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

type Analyzer struct {
	program       *ast.Program
	errors        []error
	current       *SymbolTable
	requireReturn bool
	returnTypes   []types.Type
}

func New(program *ast.Program) *Analyzer {
	return &Analyzer{
		program:       program,
		errors:        []error{},
		current:       NewSymbolTable(),
		requireReturn: false,
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
	case *ast.DeclareStatement:
		a.analyzeDeclareStatement(s)
	case *ast.ExpressionStatement:
		expr := s.Expression
		a.analyzeExpression(expr)
	case *ast.FunctionStatement:
		a.analyzeFunctionStatement(s)
	default:
		a.error(fmt.Sprintf("evaluateStatement got unexpected statement: %T", stmt))
	}
}

func (a *Analyzer) analyzeExpression(exp ast.Expression) types.Type {
	if exp == nil {
		a.error("analyzeExpression called with nil expression")
		return types.InvalidType
	}

	switch e := exp.(type) {
	case *ast.IntegerLiteral:
		return types.Int64Type
	case *ast.Boolean:
		return types.BoolType
	case *ast.Identifier:
		sym, ok := a.current.Get(e.Value)

		if !ok {
			a.error("undefined identifier: " + e.Value)
			return types.InvalidType
		}
		t := sym.Type()
		return t
	case *ast.InfixExpression:
		return a.analyzeInfixExpression(e)
	case *ast.CallExpression:
		return a.analyzeCallExpression(e)
	case *ast.IfExpression:
		return a.analyzeIfStatement(e)
	}

	a.error(fmt.Sprintf("analyzeExpression: unknown expression type: %T", exp))
	return types.InvalidType
}

func (a *Analyzer) analyzeDeclareStatement(s *ast.DeclareStatement) {
	isTypeDefined := s.Type != nil

	var right types.Type
	switch rs := s.Value.(type) {
	case *ast.IfExpression:
		right = a.analyzeIfExpressionWithValue(rs)
	default:
		right = a.analyzeExpression(rs)
	}

	if types.IsInvalid(right) {
		return
	}

	if isTypeDefined {
		if !types.IsTypesEqual(s.Type, right) {
			a.error(
				fmt.Sprintf(
					"declaration value type (%s) does not match declared type (%s)",
					right.Name(), s.Type.Name(),
				),
			)
			return
		}
	} else {
		s.SetInferredType(right)
	}

	if a.current.ExistsInCurrentScope(s.Name.Value) {
		a.error(fmt.Sprintf(
			"variable '%s' already declared in current scope",
			s.Name.Value,
		))
		return
	}

	a.current.Set(s.Name.Value, &DeclareSymbol{
		name: s.Name.Value,
		typ:  s.GetType(),
		mut:  s.Mutable,
	})
}

func (a *Analyzer) analyzeFunctionStatement(stmt *ast.FunctionStatement) {
	params := make([]BasicSymbol, len(stmt.Parameters))
	if len(stmt.Parameters) > 0 {
		for i, p := range stmt.Parameters {
			sym := BasicSymbol{name: p.Name.Value, typ: p.Type}
			params[i] = sym
		}
	}

	returnTypes := make([]types.Type, len(stmt.ReturnTypes))
	if len(stmt.ReturnTypes) > 0 {
		for i, rt := range stmt.ReturnTypes {
			returnTypes[i] = rt.Type
		}
	}

	a.current.Set(stmt.Name.Value, &FunctionSymbol{
		name:        stmt.Name.Value,
		typ:         types.FunctionType,
		params:      params,
		returnTypes: returnTypes,
	})

	a.enterScope()
	for _, p := range stmt.Parameters {
		a.current.Set(p.Name.Value, &BasicSymbol{name: p.Name.Value, typ: p.Type})
	}

	bodyStmts := stmt.Body.Statements
	for _, bodyStmt := range bodyStmts {
		returnStmt, ok := bodyStmt.(*ast.ReturnStatement)
		if ok {
			rtLen := len(stmt.ReturnTypes)
			rvLen := len(returnStmt.ReturnValues)

			if rtLen == 0 && rvLen > 0 {
				// log.Printf("rvLen: %+v", returnStmt.ReturnValues)
				a.error(fmt.Sprintf("unexpected return with values found in function %s", stmt.Name.Value))
				continue
			}

			if rvLen == 0 && rtLen > 0 {
				a.error(fmt.Sprintf("return statement with no values found in function %s", stmt.Name.Value))
				continue
			}

			if rtLen != rvLen {
				a.error(fmt.Sprintf("wrong number of return values: expected %d, got %d", rtLen, rvLen))
				continue
			}

			// return types expected comparison
			for i, retType := range stmt.ReturnTypes {
				retValue := a.analyzeExpression(returnStmt.ReturnValues[i])
				if !types.IsTypesEqual(retValue, retType.Type) {
					a.error(fmt.Sprintf("analyzeFunctionStatement: invalid return type: %s, expected: %s", retValue, retType.Type))
					continue
				}
			}
		} else {
			a.checkStatement(bodyStmt)
		}
	}

	a.exitScope()
}

func (a *Analyzer) analyzeInfixExpression(e *ast.InfixExpression) types.Type {
	left := a.analyzeExpression(e.Left)
	right := a.analyzeExpression(e.Right)

	if types.IsInvalid(left) || types.IsInvalid(right) {
		return types.InvalidType
	}

	switch e.Operator {
	case "+", "-", "*", "/":
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
	case "<", ">", "==", "!=":
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

		return types.BoolType
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

func (a *Analyzer) analyzeIfStatement(e *ast.IfExpression) types.Type {
    cond := a.analyzeExpression(e.Condition)
    if !types.IsBoolean(cond) {
        a.error("if condition must be boolean")
        return types.InvalidType
    }

    a.enterScope()
    for _, stmt := range e.Body.Statements {
        a.checkStatement(stmt)
    }
    a.exitScope()

    if e.Else != nil {
        switch s := e.Else.(type) {
        case *ast.IfExpression:
            a.analyzeIfStatement(s)
        case *ast.BlockExpression:
            a.enterScope()
            for _, stmt := range s.Statements {
                a.checkStatement(stmt)
            }
            a.exitScope()
        default:
            a.error("unexpected else structure")
        }
    }

    return types.InvalidType
}

func (a *Analyzer) analyzeIfExpressionWithValue(e *ast.IfExpression) types.Type {
	var gatherTypes func(ifst *ast.IfExpression) []types.Type
	
	gatherTypes = func(ifst *ast.IfExpression) []types.Type {
		returnTypes := []types.Type{}

		if !a.isIfConditionBoolean(ifst) {
			returnTypes = append(returnTypes, types.InvalidType)
			return returnTypes
		}

		block := ifst.Body
		if len(block.Statements) == 0 {
			a.error("if expression body cannot be empty when used as a value")
			return []types.Type{types.InvalidType}
		}

		a.enterScope()
		for i, stmt := range block.Statements {
			switch s := stmt.(type) {
			case *ast.ExpressionStatement:
				t := a.analyzeExpression(s.Expression)
				if i == len(block.Statements) - 1 {
					returnTypes = append(returnTypes, t)
				}
			default:
				if i == len(block.Statements) - 1 {
					a.error(fmt.Sprintf("last statement of if is not an expression. Got=%T", s))
					returnTypes = append(returnTypes, types.InvalidType)
					a.exitScope()
					return returnTypes
				}

				a.checkStatement(s)
			}
		}

		a.exitScope()

		if ifst.Else != nil {
			switch s := ifst.Else.(type) {
			case *ast.IfExpression:
				secondaryReturnTypes := gatherTypes(s)
				for _, srt := range secondaryReturnTypes {
					returnTypes = append(returnTypes, srt)
				}
			case *ast.BlockExpression:
				a.enterScope()
				defer a.exitScope()
				stmts := s.Statements
				if len(stmts) == 0 {
					a.error("else block cannot be empty when if is used as value")
					return append(returnTypes, types.InvalidType)
				}
				lastStmt := stmts[len(stmts)-1]
				expStmt, ok := lastStmt.(*ast.ExpressionStatement)
				if !ok {
					a.error("last statement in else must be an expression")
					return append(returnTypes, types.InvalidType)
				}
				rt := a.analyzeExpression(expStmt.Expression)
				returnTypes = append(returnTypes, rt)
			default:
				a.error("unexpected else structure")
				returnTypes = append(returnTypes, types.InvalidType)
			}
		}


		return returnTypes
	}

	typesFound := gatherTypes(e)
	if len(typesFound) == 0 {
		a.error("expected If expression to return type, found no evaluated type expression")
		return types.InvalidType
	}

	var lastType types.Type 
	for _, t := range typesFound {
		if lastType == nil { 
			lastType = t
		} else {
			if types.IsInvalid(t) {
				a.error("Invalid type found in if else expressions")
				return types.InvalidType
			} else if !types.IsTypesEqual(lastType, t) {
				a.error(fmt.Sprintf("types from each if else expressions do not match got=%T and %T", lastType.Name(), t.Name()))
				return types.InvalidType
			}
		}
	}

	return lastType
}

func (a *Analyzer) isIfConditionBoolean(ifSt *ast.IfExpression) bool {
	t := a.analyzeExpression(ifSt.Condition)
	isBool := types.IsBoolean(t)
	if !isBool {
		a.error(fmt.Sprintf("if condition must evaluate to boolean. Got=%s", t.Name()))
	}

	return isBool
}

func (a *Analyzer) enterScope() {
	a.current = NewEnclosedSymbolTable(a.current)
}

func (a *Analyzer) exitScope() {
	a.current = a.current.outer
}

func (a *Analyzer) error(err string) {
	a.errors = append(a.errors, fmt.Errorf("%s", err))
}
