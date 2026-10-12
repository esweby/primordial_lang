package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/esweby/primordial_lang/evaluator"
	"github.com/esweby/primordial_lang/inference"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/resolver"
	"github.com/esweby/primordial_lang/semantic"
)

const (
	PROMPT              = ">> "
	CONTINUATION_PROMPT = ".. "
)

// session holds the state that carries over between submissions. It is only
// updated when a submission passes every phase.
type session struct {
	scope *resolver.Scope
	env   *object.Environment
}

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	state := &session{env: object.NewEnvironment()}
	var submission strings.Builder

	for {
		if submission.Len() == 0 {
			fmt.Fprint(out, PROMPT)
		} else {
			fmt.Fprint(out, CONTINUATION_PROMPT)
		}

		scanned := scanner.Scan()
		if !scanned {
			if strings.TrimSpace(submission.String()) != "" {
				evaluateSubmission(submission.String(), out, state)
			}
			return
		}

		line := scanner.Text()
		if submission.Len() == 0 && strings.TrimSpace(line) == "" {
			continue
		}
		if submission.Len() > 0 {
			submission.WriteByte('\n')
		}
		submission.WriteString(line)

		if !submissionComplete(submission.String()) {
			continue
		}

		evaluateSubmission(submission.String(), out, state)
		submission.Reset()
	}
}

func evaluateSubmission(source string, out io.Writer, state *session) {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		printErrors(out, p.Errors())
		return
	}

	scope, resolveErrs := resolver.NewContinuing(program, state.scope).Resolve()
	if len(resolveErrs) > 0 {
		printErrors(out, resolveErrs)
		return
	}

	if inferErrs := inference.New(program).Infer(); len(inferErrs) > 0 {
		printErrors(out, inferErrs)
		return
	}

	if analyzeErrs := semantic.New(program).Analyze(); len(analyzeErrs) > 0 {
		printErrors(out, analyzeErrs)
		return
	}

	state.scope = scope
	evaluator := evaluator.New(program, state.env)
	evaluated := evaluator.Evaluate()
	if evaluated != nil {
		io.WriteString(out, evaluated.Inspect())
		io.WriteString(out, "\n")
	}
}

func submissionComplete(source string) bool {
	stack := make([]byte, 0, 8)
	inString := false

	for i := 0; i < len(source); i++ {
		character := source[i]
		if inString {
			if character == '"' {
				inString = false
			}
			continue
		}

		switch character {
		case '"':
			inString = true
		case '{', '(', '[':
			stack = append(stack, character)
		case '}', ')', ']':
			if len(stack) == 0 || !matchingDelimiter(stack[len(stack)-1], character) {
				return true
			}
			stack = stack[:len(stack)-1]
		}
	}

	return !inString && len(stack) == 0
}

func matchingDelimiter(open, close byte) bool {
	return open == '{' && close == '}' ||
		open == '(' && close == ')' ||
		open == '[' && close == ']'
}

func printErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
