package resolver

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
)

func (r *Resolver) resolveExpression(exp ast.Expression) {
	switch e := exp.(type) {
	case nil:
		return
	case *ast.IntegerLiteral, *ast.StringLiteral, *ast.Boolean:
		return
	case *ast.Identifier:
		r.lookup(e)
	case *ast.PrefixExpression:
		r.resolveExpression(e.Right)
	case *ast.InfixExpression:
		r.resolveExpression(e.Left)
		r.resolveExpression(e.Right)
	case *ast.CallExpression:
		r.resolveExpression(e.Function)
		for _, arg := range e.Arguments {
			r.resolveExpression(arg)
		}
	case *ast.MemberExpression:
		// Only the receiver is a name in scope; the member is resolved
		// against the receiver's type in a later phase.
		r.resolveExpression(e.Receiver)
	case *ast.IndexExpression:
		r.resolveExpression(e.Left)
		r.resolveExpression(e.Index)
	case *ast.IfExpression:
		r.resolveIfExpression(e)
	case *ast.BlockExpression:
		r.resolveBlock(e)
	case *ast.ForLoop:
		r.resolveForLoop(e)
	case *ast.FunctionLiteral:
		r.resolveFunctionLiteral(e)
	case *ast.ArrayLiteral:
		r.resolveType(e.Type, e.Token)
		r.resolveExpressions(e.Elements)
	case *ast.SliceLiteral:
		r.resolveType(e.Type, e.Token)
		r.resolveExpressions(e.Elements)
	case *ast.MapLiteral:
		r.resolveType(e.Type, e.Token)
		for _, pair := range e.Pairs {
			r.resolveExpression(pair.Key)
			r.resolveExpression(pair.Value)
		}
	case *ast.StructLiteral:
		r.resolveStructLiteral(e)
	case *ast.TupleTargetExpression:
		for _, name := range e.Names {
			r.lookup(name)
		}
	default:
		r.errors = append(r.errors, fmt.Sprintf("resolver received unexpected expression: %T", exp))
	}
}

func (r *Resolver) resolveExpressions(exps []ast.Expression) {
	for _, exp := range exps {
		r.resolveExpression(exp)
	}
}

func (r *Resolver) resolveIfExpression(e *ast.IfExpression) {
	r.resolveExpression(e.Condition)
	r.resolveBlock(e.Body)
	r.resolveExpression(e.Else)
}

func (r *Resolver) resolveStructLiteral(e *ast.StructLiteral) {
	if sym, ok := r.lookup(e.Name); ok && !sym.IsType() {
		r.errorAt(e.Name.Token, "%s is not a struct type", e.Name.Value)
	}

	// Field names are members of the struct; only the values are resolved.
	// Shorthand fields use the field name identifier as their value.
	for _, field := range e.Fields {
		r.resolveExpression(field.Value)
	}
}

func (r *Resolver) resolveFunctionLiteral(fnLit *ast.FunctionLiteral) *Scope {
	body := r.resolveFunction(fnLit, fnLit.Token, fnLit.Parameters, fnLit.ReturnTypes, fnLit.Body, nil)
	fnLit.Scope = body
	return body
}

// resolveForLoop opens one scope for the controller and the body together,
// so loop variables are visible in the body.
func (r *Resolver) resolveForLoop(loop *ast.ForLoop) {
	loop.Scope = r.enterScope(loop)
	defer r.exitScope()

	switch c := loop.Controller.(type) {
	case *ast.Infinite:
	case *ast.While:
		r.resolveExpression(c.Condition)
	case *ast.Constructed:
		r.resolveStatement(c.Initializer)
		r.resolveExpression(c.Condition)
		r.resolveStatement(c.Iterator)
	case *ast.Range:
		r.resolveExpression(c.Iterable)
		if len(c.Variables) < 1 || len(c.Variables) > 2 {
			r.errorAt(c.Token, "range loop requires 1 or 2 variables, got %d", len(c.Variables))
		}
		for _, variable := range c.Variables {
			r.declare(variable, RangeSymbol, c)
		}
	}

	r.loopDepth++
	if loop.Label != nil {
		r.loopLabels = append(r.loopLabels, loop.Label.Value)
	}
	defer func() {
		r.loopDepth--
		if loop.Label != nil {
			r.loopLabels = r.loopLabels[:len(r.loopLabels)-1]
		}
	}()

	if loop.Body != nil {
		loop.Body.Scope = loop.Scope
		r.resolveStatements(loop.Body.Statements)
	}
}
