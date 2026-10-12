package resolver_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/resolver"
	"github.com/esweby/primordial_lang/semantic"
)

var rawString = regexp.MustCompile("`([^`]*)`")

// knownDifferences are programs semantic accepts only because it does not
// check them, which the resolver correctly rejects.
var knownDifferences = map[string]string{
	// `int` is in the specification but not in the types package; semantic
	// does not check the key and value types of map annotations.
	`x := map[string]int{}; x["key"] = 42;`: "unknown type: int",
}

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
			_, errs := resolver.New(parser.New(lexer.New(input)).ParseProgram()).Resolve()
			if expected, known := knownDifferences[input]; known {
				if len(errs) != 1 || !strings.Contains(errs[0], expected) {
					t.Errorf("%s: expected known difference %q, got %v", file, expected, errs)
				}
				continue
			}
			if len(errs) > 0 {
				t.Errorf("%s: resolver rejected a program semantic accepts:\n%s\nerrors: %v", file, input, errs)
			}
		}
	}

	if checked == 0 {
		t.Fatal("no programs found in the corpus")
	}
	t.Logf("resolved %d programs", checked)
}
