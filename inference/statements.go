package inference

import (
	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

func (in *Inferrer) inferStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		in.inferExpressionStatement(s)
	case *ast.DeclareStatement:
		in.inferDeclareStatement(s)
	case *ast.TupleDeclareStatement:
		in.inferTupleDeclareStatement(s)
	case *ast.AssignStatement:
		in.inferAssignStatement(s)
	case *ast.TupleAssignStatement:
		in.inferTupleAssignStatement(s)
	case *ast.ReturnStatement:
		in.inferReturnStatement(s)
	case *ast.FunctionStatement:
		in.inferFunctionStatement(s)
	case *ast.StructStatement:
		in.inferStructStatement(s)
	case *ast.BreakStatement:
		in.inferBreakStatement(s)
	case *ast.ContinueStatement:
	}
}

func (in *Inferrer) inferStatements(stmts []ast.Statement) {
	for _, stmt := range stmts {
		in.inferStatement(stmt)
	}
}

// inferExpressionStatement infers an expression whose value is unused. If and
// for are statements here, so their branches and breaks need not agree.
func (in *Inferrer) inferExpressionStatement(stmt *ast.ExpressionStatement) {
	switch e := stmt.Expression.(type) {
	case *ast.IfExpression:
		in.inferIfExpression(e, nil, false)
	default:
		in.defaultUntyped(e, in.inferExpression(e, nil))
	}
}

func (in *Inferrer) inferDeclareStatement(stmt *ast.DeclareStatement) {
	in.resolveNamed(stmt.Type)

	// The name is typed before the body so a function literal can call itself.
	if fnLit, ok := stmt.Value.(*ast.FunctionLiteral); ok {
		signature := in.functionSignature(fnLit.Parameters, fnLit.ReturnTypes)
		stmt.Name.SetResolvedType(signature)
		if stmt.Type == nil {
			stmt.SetInferredType(signature)
		}
		in.inferFunctionLiteral(fnLit)
		return
	}

	if stmt.Type != nil {
		in.inferValue(stmt.Value, stmt.Type)
		stmt.Name.SetResolvedType(stmt.Type)
		return
	}

	typ := in.inferDefault(stmt.Value)
	stmt.SetInferredType(typ)
	stmt.Name.SetResolvedType(typ)
}

func (in *Inferrer) inferTupleDeclareStatement(stmt *ast.TupleDeclareStatement) {
	elements, _ := types.UnwrapTuple(in.inferDefault(stmt.Value))
	for i, name := range stmt.Names {
		if i < len(elements) {
			name.SetResolvedType(elements[i])
		} else {
			name.SetResolvedType(types.InvalidType)
		}
	}
}

func (in *Inferrer) inferAssignStatement(stmt *ast.AssignStatement) {
	var target types.Type
	if stmt.Name != nil {
		target = in.inferExpression(stmt.Name, nil)
	} else {
		target = in.inferExpression(stmt.Target, nil)
	}

	if types.IsInvalid(target) {
		in.inferDefault(stmt.Value)
		return
	}
	in.inferValue(stmt.Value, target)
}

func (in *Inferrer) inferTupleAssignStatement(stmt *ast.TupleAssignStatement) {
	elements, _ := types.UnwrapTuple(in.inferDefault(stmt.Value))
	for i, name := range stmt.Names {
		if name.Value == "_" {
			continue
		}
		target := in.inferExpression(name, nil)
		if i < len(elements) && types.IsInteger(target) && types.IsInteger(elements[i]) &&
			!types.IsTypesEqual(target, elements[i]) {
			in.errorAt(name, "cannot use %s value as %s", elements[i].Name(), target.Name())
		}
	}
}

func (in *Inferrer) inferReturnStatement(stmt *ast.ReturnStatement) {
	for i, value := range stmt.ReturnValues {
		if i < len(in.returnTypes) {
			in.inferValue(value, in.returnTypes[i])
		} else {
			in.inferDefault(value)
		}
	}
}

func (in *Inferrer) inferFunctionStatement(stmt *ast.FunctionStatement) {
	signature := in.functionSignature(stmt.Parameters, stmt.ReturnTypes)
	stmt.Name.SetResolvedType(signature)
	in.inferFunctionBody(stmt.Parameters, signature, stmt.Body)
}

func (in *Inferrer) inferBreakStatement(stmt *ast.BreakStatement) {
	if stmt.Value == nil {
		return
	}

	loop := in.breakTarget(stmt)
	if loop == nil {
		in.inferDefault(stmt.Value)
		return
	}

	if loop.expected != nil {
		in.inferValue(stmt.Value, loop.expected)
	} else {
		in.inferExpression(stmt.Value, nil)
	}
	loop.values = append(loop.values, stmt.Value)
}

// breakTarget finds the loop a break exits: the innermost loop, or the
// enclosing loop with a matching label.
func (in *Inferrer) breakTarget(stmt *ast.BreakStatement) *loopContext {
	for i := len(in.loops) - 1; i >= 0; i-- {
		loop := in.loops[i]
		if stmt.Label == nil || (loop.loop.Label != nil && loop.loop.Label.Value == stmt.Label.Value) {
			return loop
		}
	}
	return nil
}

// functionSignature builds the type of a function from its declaration.
func (in *Inferrer) functionSignature(params []*ast.Parameter, returns []*ast.ReturnType) *types.Function {
	paramTypes := make([]types.Type, len(params))
	for i, param := range params {
		in.resolveNamed(param.Type)
		paramTypes[i] = param.Type
	}

	returnTypes := make([]types.Type, len(returns))
	for i, ret := range returns {
		in.resolveNamed(ret.Type)
		returnTypes[i] = ret.Type
	}

	return types.NewFunction(paramTypes, returnTypes)
}

// inferFunctionBody types the parameters and infers the body with the
// function's return types. Loops do not cross a function boundary.
func (in *Inferrer) inferFunctionBody(params []*ast.Parameter, signature *types.Function, body *ast.BlockExpression) {
	for _, param := range params {
		param.Name.SetResolvedType(param.Type)
	}

	previousReturns, previousLoops := in.returnTypes, in.loops
	in.returnTypes, in.loops = signature.ReturnTypes, nil
	defer func() {
		in.returnTypes, in.loops = previousReturns, previousLoops
	}()

	if body != nil {
		in.inferStatements(body.Statements)
	}
}

// declareStructs creates every struct type in a statement list before any
// statement is inferred, so structs can refer to each other in any order.
// Declared field types and member signatures are filled in before defaults,
// so a default value can use any struct in the list.
func (in *Inferrer) declareStructs(stmts []ast.Statement) {
	declared := []*ast.StructStatement{}
	for _, stmt := range stmts {
		if s, ok := stmt.(*ast.StructStatement); ok {
			in.declareStruct(s)
			declared = append(declared, s)
		}
	}
	for _, s := range declared {
		in.declareStructMembers(s)
	}
	for _, s := range declared {
		in.inferStructDefaults(s)
	}
}

func (in *Inferrer) declareStruct(stmt *ast.StructStatement) *types.Struct {
	st := &types.Struct{
		TypeName:      stmt.Name.Value,
		Fields:        make(map[string]types.StructField, len(stmt.Fields)),
		TypeFunctions: make(map[string]types.MemberDefinition, len(stmt.TypeFunctions)),
		Methods:       make(map[string]types.MemberDefinition),
	}
	in.structs[stmt] = st
	stmt.Name.SetResolvedType(st)
	return st
}

// declareStructMembers records fields with a declared type and the
// signatures of type functions and methods. Duplicates are left for the
// analyzer; the first declaration wins.
func (in *Inferrer) declareStructMembers(stmt *ast.StructStatement) {
	st := in.structs[stmt]

	for _, field := range stmt.Fields {
		if field.Type == nil {
			continue
		}
		in.resolveNamed(field.Type)
		if _, duplicate := st.Fields[field.Name.Value]; !duplicate {
			st.Fields[field.Name.Value] = types.StructField{
				Name:       field.Name.Value,
				Type:       field.Type,
				HasDefault: field.Value != nil,
				Public:     field.Public,
			}
		}
	}

	for _, function := range stmt.TypeFunctions {
		if _, duplicate := st.TypeFunctions[function.Name.Value]; !duplicate {
			st.TypeFunctions[function.Name.Value] = in.memberSignature(function)
		}
	}

	if stmt.Impl != nil {
		for _, method := range stmt.Impl.Methods {
			if _, duplicate := st.Methods[method.Name.Value]; !duplicate {
				st.Methods[method.Name.Value] = in.memberSignature(method)
			}
		}
	}
}

func (in *Inferrer) memberSignature(function *ast.FunctionStatement) types.MemberDefinition {
	signature := in.functionSignature(function.Parameters, function.ReturnTypes)
	function.Name.SetResolvedType(signature)
	return types.MemberDefinition{
		Name:        function.Name.Value,
		Kind:        types.MemberMethod,
		Parameters:  signature.ParamTypes,
		ReturnTypes: signature.ReturnTypes,
		Public:      function.Public,
	}
}

// inferStructDefaults infers default values, giving fields without a declared
// type the type of their default.
func (in *Inferrer) inferStructDefaults(stmt *ast.StructStatement) {
	st := in.structs[stmt]

	for _, field := range stmt.Fields {
		if field.Value == nil {
			continue
		}
		if field.Type != nil {
			in.inferValue(field.Value, field.Type)
			continue
		}

		typ := in.inferDefault(field.Value)
		field.SetInferredType(typ)
		if _, duplicate := st.Fields[field.Name.Value]; !duplicate {
			st.Fields[field.Name.Value] = types.StructField{
				Name:       field.Name.Value,
				Type:       typ,
				HasDefault: true,
				Public:     field.Public,
			}
		}
	}
}

func (in *Inferrer) inferStructStatement(stmt *ast.StructStatement) {
	// Structs outside the top level are declared where they appear.
	if _, declared := in.structs[stmt]; !declared {
		in.declareStruct(stmt)
		in.declareStructMembers(stmt)
		in.inferStructDefaults(stmt)
	}

	for _, function := range stmt.TypeFunctions {
		in.inferFunctionBody(function.Parameters, in.signatureOf(function), function.Body)
	}
	if stmt.Impl != nil {
		for _, method := range stmt.Impl.Methods {
			in.inferFunctionBody(method.Parameters, in.signatureOf(method), method.Body)
		}
	}
}

func (in *Inferrer) signatureOf(function *ast.FunctionStatement) *types.Function {
	if signature, ok := function.Name.GetResolvedType().(*types.Function); ok {
		return signature
	}
	return in.functionSignature(function.Parameters, function.ReturnTypes)
}
