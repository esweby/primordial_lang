package resolver_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/resolver"
	"github.com/esweby/primordial_lang/semantic"
)

var rawString = regexp.MustCompile("`([^`]*)`")

// TestResolver_AcceptsSemanticCorpus resolves every program in the semantic
// and evaluator test files that parses and passes semantic analysis. Any such
// program must also resolve cleanly, otherwise the resolver is rejecting
// valid code.
func TestResolver_AcceptsSemanticCorpus(t *testing.T) {
	files, err := filepath.Glob("../semantic/*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	evaluatorFiles, err := filepath.Glob("../evaluator/*_test.go")
	if err != nil {
		t.Fatal(err)
	}
	files = append(files, evaluatorFiles...)

	checked := 0
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}

		for _, match := range rawString.FindAllStringSubmatch(string(src), -1) {
			input := match[1]
			p := parser.New(lexer.New(input))
			program := p.ParseProgram()
			if len(p.Errors()) > 0 || len(program.Statements) == 0 {
				continue
			}
			if errs := semantic.NewSemanticAnalyzer(program, semantic.NewSymbolTable()).Analyze(); len(errs) > 0 {
				continue
			}

			// semantic writes inferred types onto the AST, so resolve a fresh parse.
			checked++
			if _, errs := resolver.New(parser.New(lexer.New(input)).ParseProgram()).Resolve(); len(errs) > 0 {
				t.Errorf("%s: resolver rejected a program semantic accepts:\n%s\nerrors: %v", file, input, errs)
			}
		}
	}

	if checked == 0 {
		t.Fatal("no programs found in the corpus")
	}
	t.Logf("resolved %d programs", checked)
}
