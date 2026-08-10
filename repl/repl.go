package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/esweby/primordial_lang/evaluator"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/semantic"
)

const (
	PROMPT              = ">> "
	CONTINUATION_PROMPT = ".. "
)

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	symbols := semantic.NewSymbolTable()
	env := object.NewEnvironment()
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
				symbols = evaluateSubmission(submission.String(), out, symbols, env)
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

		symbols = evaluateSubmission(submission.String(), out, symbols, env)
		submission.Reset()
	}
}

func evaluateSubmission(
	source string,
	out io.Writer,
	symbols *semantic.SymbolTable,
	env *object.Environment,
) *semantic.SymbolTable {
	l := lexer.New(source)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		printParserErrors(out, p.Errors())
		return symbols
	}

	sa := semantic.NewSemanticAnalyzer(program, symbols)
	saErrs := sa.Analyze()

	if len(saErrs) > 0 {
		printSaErrors(out, saErrs)
		return symbols
	}

	symbols = sa.Symbols()

	evaluated := evaluator.Eval(program, env)
	if evaluated != nil {
		io.WriteString(out, evaluated.Inspect())
		io.WriteString(out, "\n")
	}

	return symbols
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

func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}

func printSaErrors(out io.Writer, errors []error) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg.Error()+"\n")
	}
}
