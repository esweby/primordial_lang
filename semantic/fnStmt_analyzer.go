package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

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