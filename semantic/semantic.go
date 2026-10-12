package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/resolver"
	"github.com/esweby/primordial_lang/types"
)

// Analyzer checks that a resolved and inferred program follows the language's
// rules. Names are checked by the resolver and integer rules by inference, so
// the analyzer reads the types inference recorded on the AST rather than
// working them out, and only reports rules that are not about either.
type Analyzer struct {
	program *ast.Program
	errors  []string

	function      *functionContext
	currentStruct *types.Struct
	loops         []*loopContext
}

// functionContext is the function whose body is being analyzed.
type functionContext struct {
	returnTypes []types.Type
}

// loopContext collects the breaks that exit a loop.
type loopContext struct {
	loop   *ast.ForLoop
	breaks []*ast.BreakStatement
}

// New creates an Analyzer for a program that has been resolved and inferred.
func New(program *ast.Program) *Analyzer {
	return &Analyzer{program: program, errors: []string{}}
}

func (a *Analyzer) Analyze() []string {
	if a.program == nil || a.program.Scope == nil {
		a.errors = append(a.errors, "analysis requires a resolved and inferred program")
		return a.errors
	}

	for _, stmt := range a.program.Statements {
		a.analyzeStatement(stmt)
	}

	return a.errors
}

func (a *Analyzer) Errors() []string { return a.errors }

// checkAssignable reports a value whose type cannot be used where target is
// expected. Integer mismatches are left to inference, and a value of any
// function type can be used where `function` is expected.
func (a *Analyzer) checkAssignable(node ast.Node, target, actual types.Type, context string) {
	if types.IsInvalid(target) || types.IsInvalid(actual) {
		return
	}
	if types.IsInteger(target) && types.IsInteger(actual) {
		return
	}
	if target == types.FunctionType && types.IsFunction(actual) {
		return
	}
	if !types.IsTypesEqual(target, actual) {
		a.errorAt(node, "%s: expected %s, got %s", context, target.Name(), actual.Name())
	}
}

// sameType reports whether two values agree, leaving integer mismatches to
// inference.
func sameType(left, right types.Type) bool {
	if types.IsInteger(left) && types.IsInteger(right) {
		return true
	}
	return types.IsTypesEqual(left, right)
}

// variable reports whether a symbol is a variable, and whether it is mutable.
// Tuple elements are immutable; parameters, range variables and self are not
// variables.
func variable(sym *resolver.Symbol) (isVariable, mutable bool) {
	if sym == nil {
		return false, false
	}
	switch sym.Kind {
	case resolver.VariableSymbol:
		declaration, ok := sym.Decl.(*ast.DeclareStatement)
		return true, ok && declaration.Mutable
	case resolver.TupleElementSymbol:
		return true, false
	default:
		return false, false
	}
}

// blockReturns reports whether a block always returns: it contains a return
// statement, or an if whose branches all return.
func blockReturns(block *ast.BlockExpression) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.ReturnStatement:
			return true
		case *ast.ExpressionStatement:
			if ifExpr, ok := s.Expression.(*ast.IfExpression); ok && ifReturns(ifExpr) {
				return true
			}
		}
	}
	return false
}

func ifReturns(e *ast.IfExpression) bool {
	if !blockReturns(e.Body) {
		return false
	}
	switch elseBranch := e.Else.(type) {
	case *ast.BlockExpression:
		return blockReturns(elseBranch)
	case *ast.IfExpression:
		return ifReturns(elseBranch)
	default:
		return false
	}
}

func symbolOf(ident *ast.Identifier) *resolver.Symbol {
	if ident == nil {
		return nil
	}
	sym, _ := ident.Symbol.(*resolver.Symbol)
	return sym
}

func typeOf(expr ast.Expression) types.Type {
	if typed, ok := expr.(ast.TypedExpression); ok && typed.GetResolvedType() != nil {
		return typed.GetResolvedType()
	}
	return types.InvalidType
}

// underlying unwraps named types.
func underlying(typ types.Type) types.Type {
	for {
		named, ok := typ.(*types.Named)
		if !ok {
			return typ
		}
		typ = named.Underlying
	}
}

func (a *Analyzer) errorAt(node ast.Node, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if tok := ast.TokenOf(node); tok.Line > 0 {
		msg = fmt.Sprintf("%d:%d: %s", tok.Line, tok.Column, msg)
	}
	a.errors = append(a.errors, msg)
}
