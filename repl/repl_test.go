package repl

import (
	"bytes"
	"strings"
	"testing"
)

func runREPL(input string) string {
	var output bytes.Buffer
	Start(strings.NewReader(input), &output)
	return output.String()
}

func TestExistingFeaturesWorkEndToEnd(t *testing.T) {
	output := runREPL(strings.Join([]string{
		`"hello"`,
		`len("hello")`,
		`[]int64{10, 20}[1]`,
		`1 <= 2`,
		`2 >= 3`,
		`"a" == "a"`,
	}, "\n"))

	for _, expected := range []string{"hello\n", "5\n", "20\n", "true\n", "false\n"} {
		if !strings.Contains(output, expected) {
			t.Errorf("expected output to contain %q, got %q", expected, output)
		}
	}
}

func TestRuntimeFailuresReturnErrorsInsteadOfPanicking(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "division by zero",
			input:    "1 / 0\n",
			expected: "division by zero",
		},
		{
			name: "higher order arity mismatch",
			input: strings.Join([]string{
				`fn invoke(f function) { f(); }`,
				`fn needs(x int64) { x; }`,
				`invoke(needs)`,
			}, "\n"),
			expected: "ERROR: wrong number of arguments: expected 1, got 0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := runREPL(test.input)
			if !strings.Contains(output, test.expected) {
				t.Fatalf("expected output to contain %q, got %q", test.expected, output)
			}
		})
	}
}

func TestBlockDeclarationsDoNotLeak(t *testing.T) {
	output := runREPL("if (false) { ghost := 41; }\nghost + 1\n")
	if !strings.Contains(output, "undefined identifier: ghost") {
		t.Fatalf("expected a semantic undefined identifier error, got %q", output)
	}
	if strings.Contains(output, "ERROR: identifier not found") {
		t.Fatalf("symbol table and runtime environment diverged: %q", output)
	}
}

func TestMutatingMethodRequiresMutableVariable(t *testing.T) {
	immutable := runREPL("items := []int64{};\nitems.append(1);\n")
	if !strings.Contains(immutable, "cannot call mutating method append on immutable variable: items") {
		t.Fatalf("expected immutable receiver error, got %q", immutable)
	}

	mutable := runREPL("mut items := []int64{};\nitems.append(1);\nitems.length\n")
	if !strings.Contains(mutable, "1\n") {
		t.Fatalf("expected mutable slice to contain one item, got %q", mutable)
	}
}

func TestBasicStructsWorkEndToEnd(t *testing.T) {
	output := runREPL(strings.Join([]string{
		`struct Person { pub name: string; pub age: int64 = 42; }`,
		`person := Person{name: "Ada"};`,
		`person.name`,
		`person.age`,
	}, "\n"))

	if !strings.Contains(output, "Ada\n") || !strings.Contains(output, "42\n") {
		t.Fatalf("expected struct fields and defaults to evaluate, got %q", output)
	}
}

func TestMultilineStructSubmission(t *testing.T) {
	output := runREPL(strings.Join([]string{
		`struct Person {`,
		`    age: int32;`,
		`    pub name: string;`,
		``,
		`    fn new(): Person {`,
		`        return Person{`,
		`            age: 0,`,
		`            name: "",`,
		`        };`,
		`    }`,
		``,
		`    impl {`,
		`        fn setName(newName string) {`,
		`            self.name = newName;`,
		`        }`,
		``,
		`        fn getName(): string {`,
		`            return self.name;`,
		`        }`,
		`    }`,
		`}`,
		`person := Person.new();`,
		`person.setName("Ada");`,
		`person.getName();`,
	}, "\n"))

	if !strings.Contains(output, CONTINUATION_PROMPT) {
		t.Fatalf("expected continuation prompts, got %q", output)
	}
	if !strings.Contains(output, "Ada\n") {
		t.Fatalf("expected multiline struct to evaluate, got %q", output)
	}
	for _, unexpected := range []string{"parser", "no prefix parse function", "expected next token"} {
		if strings.Contains(output, unexpected) {
			t.Fatalf("unexpected parser failure %q in %q", unexpected, output)
		}
	}
}

func TestSubmissionCompletenessTracksDelimitersOutsideStrings(t *testing.T) {
	tests := []struct {
		source   string
		complete bool
	}{
		{source: `struct Person {`, complete: false},
		{source: "struct Person {\nname: string;\n}", complete: true},
		{source: `values := []int64{1, 2`, complete: false},
		{source: `values := []int64{1, 2}`, complete: true},
		{source: `"braces { in a string }"`, complete: true},
		{source: `fn call(`, complete: false},
		{source: `fn call(]`, complete: true},
	}

	for _, test := range tests {
		if actual := submissionComplete(test.source); actual != test.complete {
			t.Errorf("submissionComplete(%q): expected %t, got %t", test.source, test.complete, actual)
		}
	}
}
