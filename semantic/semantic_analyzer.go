package semantic

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/types"
)

// BlockResult captures the result of analyzing a block of statements.
type BlockResult struct {
	Type       types.Type
	Returns    bool
	Expression ast.Expression
}

// SemanticAnalyzer performs type checking and symbol resolution.
type SemanticAnalyzer struct {
	program            *ast.Program
	errors             []error
	current            *SymbolTable
	returnTypes        []types.Type
	currentStruct      *types.Struct
	structDeclarations map[*ast.StructStatement]*types.Struct
}

// NewSemanticAnalyzer creates a new analyzer.
func NewSemanticAnalyzer(program *ast.Program, symbols *SymbolTable) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		program:            program,
		errors:             []error{},
		current:            symbols.Clone(),
		structDeclarations: make(map[*ast.StructStatement]*types.Struct),
	}
}

func (sa *SemanticAnalyzer) Symbols() *SymbolTable {
	return sa.current
}

// Analyze runs semantic analysis on the entire program.
func (sa *SemanticAnalyzer) Analyze() []error {
	stmts := sa.program.Statements
	if len(stmts) == 0 {
		sa.error("analyzer invoked on empty program")
		return sa.errors
	}

	for _, stmt := range stmts {
		if declaration, ok := stmt.(*ast.StructStatement); ok {
			sa.predeclareStruct(declaration)
		}
	}

	for _, stmt := range stmts {
		statementType := sa.analyzeStatement(stmt)
		if expressionStatement, ok := stmt.(*ast.ExpressionStatement); ok && types.IsUntypedInteger(statementType) {
			if _, err := sa.defaultInteger(expressionStatement.Expression, statementType); err != nil {
				sa.error(err.Error())
			}
		}
	}

	return sa.errors
}

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

	case *ast.StructStatement:
		return sa.analyzeStructStatement(s)

	default:
		sa.error(fmt.Sprintf("analyzeStatement received unexpected statement: %T", stmt))
		return nil
	}
}

func (sa *SemanticAnalyzer) analyzeScopedBlock(block *ast.BlockExpression) BlockResult {
	sa.enterScope()
	defer sa.exitScope()
	return sa.analyzeBlock(block)
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
