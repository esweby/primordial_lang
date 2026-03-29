package repl

import (
	"bufio"
	"fmt"
	"io"

	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/token"
)

const PROMPT = ">>"

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Fprintf(out, PROMPT)
		
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)

		for tok := l.NextToken(); tok.Type != token.EOF; tok = l.NextToken() {
			tokName := token.GetTokenName(int(tok.Type))
			fmt.Fprintf(out, "{Type:%d/%s Literal:%s}\n", tok.Type, tokName, tok.Literal)
		}
	}
} 