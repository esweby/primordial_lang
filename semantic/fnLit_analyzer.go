package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (a *Analyzer) analyzeFunctionLiteral(ds *ast.DeclareStatement) types.Type {
	// analyze if declare statement has a type assigned of 
	// function (usually functions define signatures so maybe 
	// we need to consider the scope of that here) for now, keep it 
	// basic and just look for type function

	if ds.Type != nil {
		if !types.IsFunction(ds.Type) {
			a.error(fmt.Sprintf("Expected declared type to be FunctionType. Got=%s", ds.Type.Name()))
			return types.InvalidType
		}
	} else {
		ds.SetInferredType(types.FunctionType)
	}

	// value has already coerced the statement type to function 
	// literal so need to check right's type and go ahead and 
	// analyze behaviour 

	fnLit, _ := ds.Value.(*ast.FunctionLiteral)

	params := make([]BasicSymbol, len(fnLit.Parameters))
	if len(fnLit.Parameters) > 0 {
		for i, p := range fnLit.Parameters {
			sym := BasicSymbol{name: p.Name.Value, typ: p.Type}
			params[i] = sym
		}
	}

	returnTypes := make([]types.Type, len(fnLit.ReturnTypes))
	if len(fnLit.ReturnTypes) > 0 {
		for i, rt := range fnLit.ReturnTypes {
			returnTypes[i] = rt.Type
		}
	}

	a.current.Set(ds.Name.Value, &FunctionSymbol{
		name:        ds.Name.Value,
		typ:         types.FunctionType,
		params:      params,
		returnTypes: returnTypes,
	})

	a.enterScope()
	for _, p := range fnLit.Parameters {
		a.current.Set(p.Name.Value, &BasicSymbol{name: p.Name.Value, typ: p.Type})
	}

	bodyStmts := fnLit.Body.Statements
	for _, bodyStmt := range bodyStmts {
		returnStmt, ok := bodyStmt.(*ast.ReturnStatement)
		if ok {
			rtLen := len(fnLit.ReturnTypes)
			rvLen := len(returnStmt.ReturnValues)

			if rtLen == 0 && rvLen > 0 {
				// log.Printf("rvLen: %+v", returnStmt.ReturnValues)
				a.error(fmt.Sprintf("unexpected return with values found in function %s", ds.Name.Value))
				continue
			}

			if rvLen == 0 && rtLen > 0 {
				a.error(fmt.Sprintf("return statement with no values found in function %s", ds.Name.Value))
				continue
			}

			if rtLen != rvLen {
				a.error(fmt.Sprintf("wrong number of return values: expected %d, got %d", rtLen, rvLen))
				continue
			}

			// return types expected comparison
			for i, retType := range fnLit.ReturnTypes {
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

	return types.FunctionType
}