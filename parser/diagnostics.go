package parser

import (
	"fmt"

	"github.com/esweby/primordial_lang/token"
)

// Diagnostic is a structured syntax error. Code is stable enough for tests
// and tooling; Message is intended for people and may become more descriptive.
type Diagnostic struct {
	Phase    DiagnosticPhase
	Code     string
	Message  string
	Span     token.Span
	Expected []token.TokenType
	Found    token.Token
}

type DiagnosticPhase string

const (
	PhaseLexing  DiagnosticPhase = "lexer"
	PhaseParsing DiagnosticPhase = "parser"
)

func (d Diagnostic) Error() string {
	return fmt.Sprintf(
		"%d:%d [%s] %s",
		d.Span.Start.Line,
		d.Span.Start.Column,
		d.Code,
		d.Message,
	)
}

func (p *Parser) addDiagnostic(code, message string, found token.Token, expected ...token.TokenType) {
	p.diagnostics = append(p.diagnostics, Diagnostic{
		Phase:    PhaseParsing,
		Code:     code,
		Message:  message,
		Span:     found.Span,
		Expected: append([]token.TokenType(nil), expected...),
		Found:    found,
	})
}

func (p *Parser) addLexicalDiagnostic(found token.Token) {
	message := found.LexError
	if message == "" {
		message = "invalid token"
	}
	p.diagnostics = append(p.diagnostics, Diagnostic{
		Phase:   PhaseLexing,
		Code:    "L1001",
		Message: message,
		Span:    found.Span,
		Found:   found,
	})
}

func (p *Parser) ensureDiagnostic(
	previousCount int,
	code string,
	message string,
	found token.Token,
	expected ...token.TokenType,
) {
	if len(p.diagnostics) == previousCount {
		p.addDiagnostic(code, message, found, expected...)
	}
}

func describeToken(tok token.Token) string {
	if tok.Type == token.EOF {
		return "end of input"
	}
	if tok.Literal != "" {
		return fmt.Sprintf("%s %q", token.GetTokenName(int(tok.Type)), tok.Literal)
	}
	return token.GetTokenName(int(tok.Type))
}
