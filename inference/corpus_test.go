package inference_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/inference"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/resolver"
	"github.com/esweby/primordial_lang/semantic"
	"github.com/esweby/primordial_lang/types"
)

var rawString = regexp.MustCompile("`([^`]*)`")

// TestInfer_MatchesSemanticCorpus infers every program in the semantic and
// evaluator test files that passes semantic analysis. Each must infer without
// errors and give every declaration the same type semantic does, except that
// untyped integers now default to int where semantic used int64, and function
// literals get their full signature where semantic used the generic function
// type.
func TestInfer_MatchesSemanticCorpus(t *testing.T) {
	files, err := filepath.Glob("../semantic/*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	evaluatorFiles, err := filepath.Glob("../evaluator/*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, evaluatorFiles...)

	checked, defaultChanges, signatures := 0, 0, 0
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		for _, match := range rawString.FindAllStringSubmatch(string(src), -1) {
			input := match[1]
			analyzed, ok := parse(input)
			if !ok {
				continue
			}
			if errs := semantic.NewSemanticAnalyzer(analyzed, semantic.NewSymbolTable()).Analyze(); len(errs) > 0 {
				continue
			}

			inferred, _ := parse(input)
			if _, errs := resolver.New(inferred).Resolve(); len(errs) > 0 {
				t.Errorf("%s: resolver rejected:\n%s\nerrors: %v", file, input, errs)
				continue
			}
			checked++
			if errs := inference.New(inferred).Infer(); len(errs) > 0 {
				t.Errorf("%s: inference rejected a program semantic accepts:\n%s\nerrors: %v", file, input, errs)
				continue
			}

			want, got := declarations(analyzed), declarations(inferred)
			for i := range want {
				wantType, gotType := typeName(want[i]), typeName(got[i])
				switch {
				case wantType == gotType:
				case wantType == "int64" && gotType == "int":
					defaultChanges++
				case want[i].Type == types.FunctionType && types.IsFunction(got[i].Type):
					signatures++
				default:
					t.Errorf("%s: %s has type %s, semantic gave %s, in:\n%s", file, got[i].Name.Value, gotType, wantType, input)
				}
			}
		}
	}

	if checked == 0 {
		t.Fatal("no programs found in the corpus")
	}
	t.Logf("inferred %d programs; %d declarations default to int where semantic used int64; %d function literals have full signatures",
		checked, defaultChanges, signatures)
}

func parse(input string) (*ast.Program, bool) {
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	return program, len(p.Errors()) == 0 && len(program.Statements) > 0
}

func typeName(stmt *ast.DeclareStatement) string {
	if stmt.Type == nil {
		return "<nil>"
	}
	return stmt.Type.Name()
}

// declarations collects every declare statement in source order.
func declarations(program *ast.Program) []*ast.DeclareStatement {
	found := []*ast.DeclareStatement{}

	var statements func([]ast.Statement)
	var expression func(ast.Expression)
	block := func(b *ast.BlockExpression) {
		if b != nil {
			statements(b.Statements)
		}
	}
	statement := func(stmt ast.Statement) {
		switch s := stmt.(type) {
		case *ast.DeclareStatement:
			found = append(found, s)
			expression(s.Value)
		case *ast.ExpressionStatement:
			expression(s.Expression)
		case *ast.FunctionStatement:
			block(s.Body)
		case *ast.StructStatement:
			for _, function := range s.TypeFunctions {
				block(function.Body)
			}
			if s.Impl != nil {
				for _, method := range s.Impl.Methods {
					block(method.Body)
				}
			}
		}
	}
	statements = func(stmts []ast.Statement) {
		for _, stmt := range stmts {
			statement(stmt)
		}
	}
	expression = func(expr ast.Expression) {
		switch e := expr.(type) {
		case *ast.IfExpression:
			block(e.Body)
			expression(e.Else)
		case *ast.BlockExpression:
			block(e)
		case *ast.ForLoop:
			if c, ok := e.Controller.(*ast.Constructed); ok {
				statement(c.Initializer)
			}
			block(e.Body)
		case *ast.FunctionLiteral:
			block(e.Body)
		}
	}

	statements(program.Statements)
	return found
}
