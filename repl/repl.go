package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/esweby/primordial_lang/evaluator"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/semantic"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	symbols := semantic.NewSymbolTable()
	env := object.NewEnvironment()

	for {
		fmt.Fprintf(out, PROMPT)
		
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		sa := semantic.NewSemanticAnalyzer(program, symbols)
		saErrs := sa.Analyze()

		if len(saErrs) > 0 {
			printSaErrors(out, saErrs)
			continue
		}

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
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
