package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/esweby/primordial_lang/evaluator"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/object"
	"github.com/esweby/primordial_lang/parser"
	"github.com/esweby/primordial_lang/repl"
	"github.com/esweby/primordial_lang/semantic"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "run":
			if len(os.Args) < 3 {
				fmt.Println("Usage: pri run <filename>")
				os.Exit(1)
			}
			runFile(os.Args[2])
			return
		case "repl":
			startRepl()
			return
		default:
			fmt.Printf("Unknown command: %s\n", os.Args[1])
			fmt.Println("Commands: run, repl")
			os.Exit(1)
		}
	}
	// default: start REPL
	startRepl()
}

func startRepl() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is Primordial Lang :)\n", user.Username)
	fmt.Printf("Feel free to type in commands\n")
	repl.Start(os.Stdin, os.Stdout)
}

func runFile(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	input := string(data)
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, errMsg := range p.Errors() {
			fmt.Printf("Parse error: %s\n", errMsg)
		}
		os.Exit(1)
	}

	symbols := semantic.NewSymbolTable()
	sa := semantic.NewSemanticAnalyzer(program, symbols)
	saErrs := sa.Analyze()
	if len(saErrs) > 0 {
		for _, err := range saErrs {
			fmt.Printf("Semantic error: %v\n", err)
		}
		os.Exit(1)
	}

	env := object.NewEnvironment()
	result := evaluator.Eval(program, env)

	if result != nil && result.Type() != object.VOID_OBJ {
		fmt.Println(result.Inspect())
	}

	if result != nil && result.Type() == object.ERROR_OBJ {
		os.Exit(1)
	}
}