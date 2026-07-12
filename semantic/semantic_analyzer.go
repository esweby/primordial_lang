package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

// BlockResult captures the result of analyzing a block of statements.
type BlockResult struct {
	Type    types.Type // type of the last expression (if any)
	Returns bool       // true if the block unconditionally returns
}

// SemanticAnalyzer performs type checking and symbol resolution.
type SemanticAnalyzer struct {
	program        *ast.Program
	errors         []error
	current        *SymbolTable
	returnTypes    []types.Type
}

// NewSemanticAnalyzer creates a new analyzer.
func NewSemanticAnalyzer(program *ast.Program, symbols *SymbolTable) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		program:        program,
		errors:         []error{},
		current:        symbols.Clone(),
	}
}

func (sa *SemanticAnalyzer) Symbols() *SymbolTable {
	return sa.current
}

// Analyze runs the semantic analysis on the entire program.
func (sa *SemanticAnalyzer) Analyze() []error {
	stmts := sa.program.Statements
	if len(stmts) == 0 {
		sa.error("analyzer invoked on empty program")
		return sa.errors
	}

	for _, stmt := range stmts {
		sa.analyzeStatement(stmt)
	}

	return sa.errors
}

// ============================================================================
// Statement Analysis
// ============================================================================

// analyzeStatement analyzes a statement and returns its type if it yields a value,
// or nil otherwise. It also handles control‑flow (e.g., return statements) via the
// returned BlockResult when called from analyzeBlock.
func (sa *SemanticAnalyzer) analyzeStatement(stmt ast.Statement) types.Type {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		// If the expression is an if expression, analyze it as a statement (no value expected).
		if ifExpr, ok := s.Expression.(*ast.IfExpression); ok {
			return sa.analyzeIfExpression(ifExpr, false)
		}
		return sa.analyzeExpression(s.Expression)

	case *ast.DeclareStatement:
		return sa.analyzeDeclareStatement(s)

	case *ast.TupleDeclareStatement:
		return sa.analyzeTupleDeclareStatement(s)

	case *ast.FunctionStatement:
		sa.analyzeFunctionStatement(s)
		return nil

	case *ast.ReturnStatement:
		sa.analyzeReturnStatement(s)
		return nil

	case *ast.AssignStatement:
		return sa.analyzeAssignmentStatement(s)

	case *ast.TupleAssignStatement:
		return sa.analyzeTupleAssignmentStatement(s)

	default:
		sa.error(fmt.Sprintf("analyzeStatement received unexpected statement: %T", stmt))
		return nil
	}
}

func (sa *SemanticAnalyzer) analyzeTupleDeclareStatement(stmt *ast.TupleDeclareStatement) types.Type {
	rhsType := sa.analyzeExpression(stmt.Value)
	elementTypes, ok := types.UnwrapTuple(rhsType)
	if !ok {
		sa.error(fmt.Sprintf("tuple declaration requires a tuple value, got %s", rhsType.Name()))
		return types.InvalidType
	}

	if len(stmt.Names) != len(elementTypes) {
		sa.error(fmt.Sprintf("tuple declaration arity mismatch: expected %d names, got %d",
			len(elementTypes), len(stmt.Names)))
		return types.InvalidType
	}

	seen := make(map[string]struct{})
	for _, name := range stmt.Names {
		if name.Value == "_" {
			continue
		}
		if _, duplicate := seen[name.Value]; duplicate {
			sa.error(fmt.Sprintf("variable '%s' appears more than once in tuple declaration", name.Value))
			return types.InvalidType
		}
		seen[name.Value] = struct{}{}

		if sa.current.ExistsInCurrentScope(name.Value) {
			sa.error(fmt.Sprintf("variable '%s' already declared in this scope", name.Value))
			return types.InvalidType
		}
	}

	for i, name := range stmt.Names {
		if name.Value == "_" {
			continue
		}
		sa.current.Set(name.Value, &DeclareSymbol{
			name: name.Value,
			typ:  elementTypes[i],
			mut:  false,
		})
	}

	return rhsType
}

func (sa *SemanticAnalyzer) analyzeTupleAssignmentStatement(stmt *ast.TupleAssignStatement) types.Type {
	rhsType := sa.analyzeExpression(stmt.Value)
	elementTypes, ok := types.UnwrapTuple(rhsType)
	if !ok {
		sa.error(fmt.Sprintf("tuple assignment requires a tuple value, got %s", rhsType.Name()))
		return types.InvalidType
	}

	if len(stmt.Names) != len(elementTypes) {
		sa.error(fmt.Sprintf("tuple assignment arity mismatch: expected %d names, got %d",
			len(elementTypes), len(stmt.Names)))
		return types.InvalidType
	}

	for i, name := range stmt.Names {
		if name.Value == "_" {
			continue
		}

		sym, found := sa.current.Get(name.Value)
		if !found {
			sa.error(fmt.Sprintf("undefined variable: %s", name.Value))
			return types.InvalidType
		}

		decl, isVariable := sym.(*DeclareSymbol)
		if !isVariable {
			sa.error(fmt.Sprintf("cannot assign to non-variable: %s", name.Value))
			return types.InvalidType
		}
		if !decl.Mutable() {
			sa.error(fmt.Sprintf("cannot assign to immutable variable: %s", name.Value))
			return types.InvalidType
		}
		if !types.IsAssignable(decl.Type(), elementTypes[i]) {
			sa.error(fmt.Sprintf("tuple assignment value %d: expected %s, got %s",
				i, decl.Type().Name(), elementTypes[i].Name()))
			return types.InvalidType
		}
	}

	return rhsType
}

// analyzeBlock analyzes a block of statements and returns the BlockResult.
// It captures the type of the last expression statement and whether the block
// unconditionally returns (i.e., ends with a return or a statement that always returns).
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

// analyzeDeclareStatement checks a variable declaration.
func (sa *SemanticAnalyzer) analyzeDeclareStatement(stmt *ast.DeclareStatement) types.Type {
	// Check redefinition
	if sa.current.ExistsInCurrentScope(stmt.Name.Value) {
		sa.error(fmt.Sprintf("variable '%s' already declared in this scope", stmt.Name.Value))
		return types.InvalidType
	}

	// Handle function literal assignment
	if _, ok := stmt.Value.(*ast.FunctionLiteral); ok {
		fnType := sa.analyzeFunctionLiteral(stmt) // analyzes body and registers symbol

		// If there is an explicit type annotation, it must be a function type.
		if stmt.Type != nil {
			if !types.IsFunction(stmt.Type) {
				sa.error(fmt.Sprintf("declaration type mismatch: expected %s, got function",
					stmt.Type.Name()))
				return types.InvalidType
			}
			// If annotation is `function`, we accept it.
		} else {
			// No annotation: infer as function type.
			stmt.SetInferredType(fnType)
		}
		return fnType
	}

	// Normal (non‑function) value analysis (existing code)
	rhsType := sa.analyzeExpression(stmt.Value)
	if types.IsInvalid(rhsType) {
		return types.InvalidType
	}

	if stmt.Type != nil && !types.IsAssignable(stmt.Type, rhsType) {
		sa.error(fmt.Sprintf("declaration type mismatch: expected %s, got %s",
			stmt.Type.Name(), rhsType.Name()))
		return types.InvalidType
	}

	if stmt.Type == nil {
		stmt.SetInferredType(rhsType)
	}

	sa.current.Set(stmt.Name.Value, &DeclareSymbol{
		name: stmt.Name.Value,
		typ:  stmt.GetType(),
		mut:  stmt.Mutable,
	})

	return stmt.GetType()
}

// analyzeFunctionStatement analyzes a function statement (e.g., fn add() { ... }).
func (sa *SemanticAnalyzer) analyzeFunctionStatement(stmt *ast.FunctionStatement) {
	// Build the function signature from the AST.
	paramTypes := make([]types.Type, len(stmt.Parameters))
	for i, p := range stmt.Parameters {
		paramTypes[i] = p.Type
	}
	returnTypes := make([]types.Type, len(stmt.ReturnTypes))
	for i, rt := range stmt.ReturnTypes {
		returnTypes[i] = rt.Type
	}

	// Register the function in the current scope with its signature.
	sa.current.Set(stmt.Name.Value, &FunctionSymbol{
		name:        stmt.Name.Value,
		typ:         types.FunctionType, // generic, but signature stored separately
		params:      stmt.Parameters,
		returnTypes: returnTypes,
	})

	// Enter function scope.
	sa.enterScope()

	// Add parameters to the new scope.
	for _, p := range stmt.Parameters {
		sa.current.Set(p.Name.Value, &BasicSymbol{
			name: p.Name.Value,
			typ:  p.Type,
		})
	}

	previousReturnTypes := sa.returnTypes
	sa.returnTypes = returnTypes
	defer func() {
		sa.returnTypes = previousReturnTypes
	}()

	// Analyze the body.
	bodyResult := sa.analyzeBlock(stmt.Body)

	// Validate return paths.
	if len(returnTypes) > 0 {
		if !bodyResult.Returns {
			sa.error(fmt.Sprintf("function '%s' declares return types but does not return on all paths",
				stmt.Name.Value))
		}
	} else {
		// Void function: if it returns unexpectedly, we catch it in analyzeReturnStatement.
		// No further action.
	}

	// Exit function scope.
	sa.exitScope()
}

// analyzeFunctionLiteral analyzes a function literal assigned to a variable.
// It registers the function in the symbol table as a FunctionSymbol.
func (sa *SemanticAnalyzer) analyzeFunctionLiteral(ds *ast.DeclareStatement) types.Type {
	fnLit, ok := ds.Value.(*ast.FunctionLiteral)
	if !ok {
		sa.error("analyzeFunctionLiteral called with non-function value")
		return types.InvalidType
	}

	// Build parameter symbols and return types
	params := make([]*ast.Parameter, len(fnLit.Parameters))
	copy(params, fnLit.Parameters)

	returnTypes := make([]types.Type, len(fnLit.ReturnTypes))
	for i, rt := range fnLit.ReturnTypes {
		returnTypes[i] = rt.Type
	}

	// Register the function symbol in the current scope
	sa.current.Set(ds.Name.Value, &FunctionSymbol{
		name:        ds.Name.Value,
		typ:         types.FunctionType,
		params:      params,
		returnTypes: returnTypes,
	})

	// Enter function scope, add parameters
	sa.enterScope()
	for _, p := range fnLit.Parameters {
		sa.current.Set(p.Name.Value, &BasicSymbol{name: p.Name.Value, typ: p.Type})
	}

	previousReturnTypes := sa.returnTypes
	sa.returnTypes = returnTypes
	defer func() {
		sa.returnTypes = previousReturnTypes
	}()

	for _, stmt := range fnLit.Body.Statements {
		sa.analyzeStatement(stmt)
	}

	sa.exitScope()
	return types.FunctionType
}

// analyzeReturnStatement checks a return statement against the current function context.
func (sa *SemanticAnalyzer) analyzeReturnStatement(stmt *ast.ReturnStatement) {
	// If we're not inside a function with return types, it's an error.
	if len(sa.returnTypes) == 0 {
		if len(stmt.ReturnValues) > 0 {
			sa.error("unexpected return with values in void function")
		}
		return
	}

	// Non-void function: must return the correct number and types.
	if len(stmt.ReturnValues) != len(sa.returnTypes) {
		sa.error(fmt.Sprintf("wrong number of return values: expected %d, got %d",
			len(sa.returnTypes), len(stmt.ReturnValues)))
		return
	}

	for i, rv := range stmt.ReturnValues {
		valType := sa.analyzeExpression(rv)
		expected := sa.returnTypes[i]
		if !types.IsAssignable(expected, valType) {
			sa.error(fmt.Sprintf("return value %d: expected %s, got %s",
				i, expected.Name(), valType.Name()))
		}
	}
}

// analyzeAssignmentStatement handles variable assignment (e.g., x = 5).
func (sa *SemanticAnalyzer) analyzeAssignmentStatement(stmt *ast.AssignStatement) types.Type {
	// The LHS must be an identifier (for now).
	ident := stmt.Name
	if ident == nil {
		// handle error
		return types.InvalidType
	}

	// Look up the variable.
	sym, ok := sa.current.Get(ident.Value)
	if !ok {
		sa.error(fmt.Sprintf("undefined variable: %s", ident.Value))
		return types.InvalidType
	}

	// Check mutability.
	declSym, ok := sym.(*DeclareSymbol)
	if !ok {
		sa.error(fmt.Sprintf("cannot assign to non-variable: %s", ident.Value))
		return types.InvalidType
	}
	if !declSym.Mutable() {
		sa.error(fmt.Sprintf("cannot assign to immutable variable: %s", ident.Value))
		return types.InvalidType
	}

	// Analyze RHS.
	rhsType := sa.analyzeExpression(stmt.Value)
	if types.IsInvalid(rhsType) {
		return types.InvalidType
	}

	// Type check.
	if !types.IsTypesEqual(declSym.Type(), rhsType) {
		sa.error(fmt.Sprintf("assignment type mismatch: expected %s, got %s",
			declSym.Type().Name(), rhsType.Name()))
		return types.InvalidType
	}

	return rhsType
}

// ============================================================================
// Expression Analysis
// ============================================================================

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

// analyzeStandaloneFunctionLiteral analyzes a function literal not attached to a declaration.
// It does not register the function in the symbol table, but it checks its body.
func (sa *SemanticAnalyzer) analyzeStandaloneFunctionLiteral(fnLit *ast.FunctionLiteral) {
	// Build signature for context.
	returnTypes := make([]types.Type, len(fnLit.ReturnTypes))
	for i, rt := range fnLit.ReturnTypes {
		returnTypes[i] = rt.Type
	}

	sa.enterScope()
	for _, p := range fnLit.Parameters {
		sa.current.Set(p.Name.Value, &BasicSymbol{
			name: p.Name.Value,
			typ:  p.Type,
		})
	}
	
	previousReturnTypes := sa.returnTypes
	sa.returnTypes = returnTypes
	defer func() {
		sa.returnTypes = previousReturnTypes
	}()

	bodyResult := sa.analyzeBlock(fnLit.Body)
	if len(returnTypes) > 0 && !bodyResult.Returns {
		sa.error("function literal declares return types but does not return on all paths")
	}

	sa.exitScope()
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

	fs, ok := sym.(*FunctionSymbol)
	if !ok {
		sa.error(fmt.Sprintf("symbol '%s' is not a function", ident.Value))
		return types.InvalidType
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

// analyzeIfExpression analyzes an if expression. If expectsValue is true,
// it requires all branches to yield a value of the same type.
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

// ============================================================================
// Helper functions
// ============================================================================

func (sa *SemanticAnalyzer) isInvalidInfixType(t types.Type) bool {
	return types.IsBoolean(t) || types.IsFunction(t) || types.IsString(t)
}

func (sa *SemanticAnalyzer) isInvalidPrefixType(t types.Type) bool {
	return !types.IsBoolean(t) && !types.IsNumeric(t)
}

func (sa *SemanticAnalyzer) enterScope() {
	sa.current = NewEnclosedSymbolTable(sa.current)
}

func (sa *SemanticAnalyzer) exitScope() {
	if sa.current.outer != nil {
		sa.current = sa.current.outer
	}
}

func (sa *SemanticAnalyzer) error(msg string) {
	sa.errors = append(sa.errors, fmt.Errorf("%s", msg))
}
