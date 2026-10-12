package inference

import (
	"fmt"
	"math/big"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/resolver"
	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

// Inferrer assigns a concrete type to every expression and declaration in a
// resolved program, and is the home of the language's integer rules.
//
// Integers follow Go: a literal is an untyped constant until something in its
// statement gives it a type, and it defaults to int when nothing does. Once a
// variable has a type it is fixed. The only errors reported here are integer
// errors; anything else that cannot be typed is given the invalid type and
// left for the analyzer to report.
type Inferrer struct {
	program *ast.Program
	errors  []string

	returnTypes []types.Type
	loops       []*loopContext
	loopBreaks  map[*ast.ForLoop][]ast.Expression
	structs     map[*ast.StructStatement]*types.Struct
}

// loopContext tracks the values a loop can break with.
type loopContext struct {
	loop     *ast.ForLoop
	expected types.Type
	values   []ast.Expression
}

// New creates an Inferrer for a program that has already been resolved.
func New(program *ast.Program) *Inferrer {
	return &Inferrer{
		program:    program,
		errors:     []string{},
		loopBreaks: make(map[*ast.ForLoop][]ast.Expression),
		structs:    make(map[*ast.StructStatement]*types.Struct),
	}
}

func (in *Inferrer) Infer() []string {
	if in.program == nil || in.program.Scope == nil {
		in.errors = append(in.errors, "inference requires a resolved program")
		return in.errors
	}

	in.declareStructs(in.program.Statements)
	for _, stmt := range in.program.Statements {
		in.inferStatement(stmt)
	}

	return in.errors
}

func (in *Inferrer) Errors() []string { return in.errors }

// assign gives expr the target type where the statement expects one. Untyped
// constants are converted to target, and integers of different types are
// reported. Other mismatches are left for the analyzer.
func (in *Inferrer) assign(expr ast.Expression, actual, target types.Type) types.Type {
	if target == nil || types.IsInvalid(actual) {
		return actual
	}
	if types.IsInvalid(target) {
		// The analyzer reports why target has no type; the value still needs one.
		return in.defaultUntyped(expr, actual)
	}

	if types.IsUntypedInteger(actual) {
		return in.convertUntyped(expr, target)
	}

	if types.IsInteger(actual) && types.IsInteger(target) && !types.IsTypesEqual(actual, target) {
		in.errorAt(expr, "cannot use %s value as %s", actual.Name(), target.Name())
	}

	return actual
}

// defaultUntyped gives an untyped constant the default integer type.
func (in *Inferrer) defaultUntyped(expr ast.Expression, actual types.Type) types.Type {
	if types.IsUntypedInteger(actual) {
		return in.convertUntyped(expr, types.DefaultIntegerType)
	}
	return actual
}

// convertUntyped gives an untyped integer expression the target type,
// reporting constants that do not fit. A target that is not an integer is an
// error, and the expression takes the default integer type instead.
func (in *Inferrer) convertUntyped(expr ast.Expression, target types.Type) types.Type {
	if !types.IsUntypedInteger(typeOf(expr)) {
		return typeOf(expr)
	}

	if !types.IsInteger(target) {
		in.errorAt(expr, "cannot use untyped integer constant as %s", target.Name())
		target = types.DefaultIntegerType
	}

	if value, ok := constantValue(expr); ok {
		if err := types.CheckRepresentable(value, target); err != nil {
			in.errorAt(expr, "%s", err.Error())
		}
		setConstantType(expr, target)
		return target
	}

	switch e := expr.(type) {
	case *ast.PrefixExpression:
		in.convertUntyped(e.Right, target)
	case *ast.InfixExpression:
		in.convertUntyped(e.Left, target)
		in.convertUntyped(e.Right, target)
	case *ast.IfExpression:
		in.convertUntyped(blockTail(e.Body), target)
		in.convertUntyped(e.Else, target)
	case *ast.BlockExpression:
		in.convertUntyped(blockTail(e), target)
	case *ast.ForLoop:
		for _, value := range in.loopBreaks[e] {
			in.convertUntyped(value, target)
		}
	}

	setType(expr, target)
	return target
}

// constantValue evaluates an integer constant expression.
func constantValue(expr ast.Expression) (*big.Int, bool) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return new(big.Int).Set(e.Value), true
	case *ast.PrefixExpression:
		value, ok := constantValue(e.Right)
		if !ok || e.Operator != "-" {
			return nil, false
		}
		return new(big.Int).Neg(value), true
	case *ast.InfixExpression:
		left, leftOK := constantValue(e.Left)
		right, rightOK := constantValue(e.Right)
		if !leftOK || !rightOK {
			return nil, false
		}
		value, ok, err := types.FoldIntegerConstant(e.Operator, left, right)
		if !ok || err != nil {
			return nil, false
		}
		return value, true
	default:
		return nil, false
	}
}

// setConstantType sets the type of every node in a constant expression.
func setConstantType(expr ast.Expression, target types.Type) {
	setType(expr, target)
	switch e := expr.(type) {
	case *ast.PrefixExpression:
		setConstantType(e.Right, target)
	case *ast.InfixExpression:
		setConstantType(e.Left, target)
		setConstantType(e.Right, target)
	}
}

// symbolType returns the type of the declaration an identifier refers to.
// Every declaring identifier carries its declaration's type.
func symbolType(ident *ast.Identifier) types.Type {
	sym, ok := ident.Symbol.(*resolver.Symbol)
	if !ok {
		return types.InvalidType
	}

	switch sym.Kind {
	case resolver.BuiltinSymbol, resolver.StructSymbol:
		// Builtins are only valid as callees, and types are not values.
		return types.InvalidType
	case resolver.SelfSymbol:
		if st, ok := sym.Decl.(*ast.StructStatement); ok {
			return typeOrInvalid(st.Name.GetResolvedType())
		}
		return types.InvalidType
	default:
		return typeOrInvalid(sym.Ident.GetResolvedType())
	}
}

// structOf returns the struct type an identifier names, if it names one.
func structOf(ident *ast.Identifier) (*types.Struct, bool) {
	sym, ok := ident.Symbol.(*resolver.Symbol)
	if !ok || sym.Kind != resolver.StructSymbol {
		return nil, false
	}
	st, ok := sym.Ident.GetResolvedType().(*types.Struct)
	return st, ok
}

// resolveNamed fills in the underlying struct of every named type inside a
// type annotation, using the link the resolver set.
func (in *Inferrer) resolveNamed(typ types.Type) {
	switch t := typ.(type) {
	case *types.Named:
		sym, ok := t.Symbol.(*resolver.Symbol)
		if !ok || sym.Ident == nil {
			return
		}
		if st, ok := sym.Ident.GetResolvedType().(*types.Struct); ok {
			t.Underlying = st
		}
	case *types.Array:
		in.resolveNamed(t.ElementType())
	case *types.Slice:
		in.resolveNamed(t.ElementType())
	case *types.Map:
		in.resolveNamed(t.Key)
		in.resolveNamed(t.Value)
	case *types.Tuple:
		for _, element := range t.Types {
			in.resolveNamed(element)
		}
	case *types.Function:
		for _, param := range t.ParamTypes {
			in.resolveNamed(param)
		}
		for _, ret := range t.ReturnTypes {
			in.resolveNamed(ret)
		}
	}
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

func typeOf(expr ast.Expression) types.Type {
	if typed, ok := expr.(ast.TypedExpression); ok {
		return typed.GetResolvedType()
	}
	return nil
}

func setType(expr ast.Expression, typ types.Type) types.Type {
	if typed, ok := expr.(ast.TypedExpression); ok {
		typed.SetResolvedType(typ)
	}
	return typ
}

func typeOrInvalid(typ types.Type) types.Type {
	if typ == nil {
		return types.InvalidType
	}
	return typ
}

func returnType(returns []types.Type) types.Type {
	switch len(returns) {
	case 0:
		return types.VoidType
	case 1:
		return returns[0]
	default:
		return &types.Tuple{Types: returns}
	}
}

func (in *Inferrer) errorAt(node ast.Node, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if tok := tokenOf(node); tok.Line > 0 {
		msg = fmt.Sprintf("%d:%d: %s", tok.Line, tok.Column, msg)
	}
	in.errors = append(in.errors, msg)
}

func tokenOf(node ast.Node) token.Token {
	switch n := node.(type) {
	case *ast.IntegerLiteral:
		return n.Token
	case *ast.Identifier:
		return n.Token
	case *ast.PrefixExpression:
		return n.Token
	case *ast.InfixExpression:
		return n.Token
	case *ast.CallExpression:
		return n.Token
	case *ast.IfExpression:
		return n.Token
	case *ast.BlockExpression:
		return n.Token
	case *ast.ForLoop:
		return n.Token
	case *ast.IndexExpression:
		return n.Token
	case *ast.MemberExpression:
		return n.Token
	case *ast.StringLiteral:
		return n.Token
	case *ast.Boolean:
		return n.Token
	case *ast.ArrayLiteral:
		return n.Token
	case *ast.SliceLiteral:
		return n.Token
	case *ast.MapLiteral:
		return n.Token
	case *ast.StructLiteral:
		return n.Token
	case *ast.FunctionLiteral:
		return n.Token
	default:
		return token.Token{}
	}
}
