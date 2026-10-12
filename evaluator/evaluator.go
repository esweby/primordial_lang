package evaluator

import (
	"fmt"
	"math/big"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/types"
)

var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// Evaluator runs a program that has passed the resolver, inference and the
// analyzer. It trusts that checking: names exist, types match and every
// integer has its concrete type, so the only errors it reports are ones that
// depend on runtime values, such as an index out of bounds or an integer
// overflow. Anything else that goes wrong is an internal error.
type Evaluator struct {
	node ast.Node
	env  *object.Environment
}

func New(node ast.Node, env *object.Environment) *Evaluator {
	return &Evaluator{
		node: node,
		env:  env,
	}
}

// In the original monkey language the Evaluator is given the first
// ast.Node created by the parser and start recursively working its
// way through the ast. Evaluate will be the public API to start
// this process.
func (e *Evaluator) Evaluate() object.Object {
	program, ok := e.node.(*ast.Program)
	if !ok {
		// no program provided
		return nil
	}
	return e.evalProgram(program.Statements, e.env)
}

func (e *Evaluator) evalProgram(stmts []ast.Statement, env *object.Environment) object.Object {
	var result object.Object
	for _, stmt := range stmts {
		result = e.eval(stmt, env)
		switch result.(type) {
		case *object.ReturnValue, *object.Error, *object.Break, *object.Continue:
			return result
		}
	}
	return result
}

func (e *Evaluator) eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case ast.Statement:
		return e.evalStatement(node, env)
	case ast.Expression:
		return e.evalExpression(node, env)
	default:
		// TODO: Diagnostics
		return nil
	}
}

// Helper functions
func (e *Evaluator) nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}

	return FALSE
}

// isTruthy reports whether a condition holds. Conditions are always bool, and
// booleans are the TRUE and FALSE singletons.
func (e *Evaluator) isTruthy(obj object.Object) bool {
	return obj == TRUE
}

func newIntegerObject(value *big.Int, integerType types.Type) *object.Integer {
	return &object.Integer{Value: new(big.Int).Set(value), IntegerType: integerType}
}

// checkedIntegerResult traps an integer that does not fit its type, rather
// than wrapping around.
func (e *Evaluator) checkedIntegerResult(value *big.Int, integerType types.Type) object.Object {
	concrete, ok := integerType.(*types.Integer)
	if !ok {
		return internalError("integer result has no concrete type")
	}
	if !concrete.CanRepresent(value) {
		return newError("integer overflow: %s is not representable as %s", value.String(), concrete.Name())
	}
	return newIntegerObject(value, integerType)
}

func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// internalError reports a state the earlier phases should have made
// impossible, so a bug there still fails loudly.
func internalError(format string, a ...interface{}) *object.Error {
	return newError("internal error: "+format, a...)
}

func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.ERROR_OBJ
}

func neutralObject(t types.Type) (object.Object, bool) {
	switch t.Kind() {
	case types.KindInteger:
		return newIntegerObject(big.NewInt(0), t), true
	case types.KindBoolean:
		return FALSE, true
	case types.KindString:
		return &object.String{Value: ""}, true
	default:
		return nil, false
	}
}
