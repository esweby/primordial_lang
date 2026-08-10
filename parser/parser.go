package parser

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/token"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

const (
	_int = iota
	LOWEST
	EQUALS
	LESSGREATER
	SUM
	PRODUCT
	PREFIX
	CALL
	DOT
	INDEX
)

var precedences = map[token.TokenType]int{
	token.EQUALS:                 EQUALS,
	token.NOT_EQUALS:             EQUALS,
	token.LTAG:                   LESSGREATER,
	token.RTAG:                   LESSGREATER,
	token.LESS_THAN_OR_EQUALS:    LESSGREATER,
	token.GREATER_THAN_OR_EQUALS: LESSGREATER,
	token.PLUS:                   SUM,
	token.MINUS:                  SUM,
	token.FORWARD_SLASH:          PRODUCT,
	token.ASTERIK:                PRODUCT,
	token.LPAREN:                 CALL,
	token.LBRACE:                 CALL,
	token.DOT:                    DOT,
	token.LBRACKET:               INDEX,
}

type Parser struct {
	l           *lexer.Lexer
	diagnostics []Diagnostic

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:           l,
		diagnostics: []Diagnostic{},
	}

	p.registerPrefixFns()
	p.registerInfixFns()

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{Statements: []ast.Statement{}}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		p.nextToken()
	}

	return program
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(tokenType token.TokenType) bool {
	return p.curToken.Type == tokenType
}

func (p *Parser) peekTokenIs(tokenType token.TokenType) bool {
	return p.peekToken.Type == tokenType
}

func (p *Parser) peekPrecedence() int {
	if precedence, ok := precedences[p.peekToken.Type]; ok {
		return precedence
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if precedence, ok := precedences[p.curToken.Type]; ok {
		return precedence
	}

	return LOWEST
}

func (p *Parser) expectPeek(tokenType token.TokenType) bool {
	if p.peekTokenIs(tokenType) {
		p.nextToken()
		return true
	}

	p.peekError(tokenType)
	return false
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) registerPrefixFns() {
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT_LITERAL, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FN, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING_LITERAL, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseLeftBracket)
}

func (p *Parser) registerInfixFns() {
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.FORWARD_SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERIK, p.parseInfixExpression)
	p.registerInfix(token.EQUALS, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQUALS, p.parseInfixExpression)
	p.registerInfix(token.LTAG, p.parseInfixExpression)
	p.registerInfix(token.RTAG, p.parseInfixExpression)
	p.registerInfix(token.LESS_THAN_OR_EQUALS, p.parseInfixExpression)
	p.registerInfix(token.GREATER_THAN_OR_EQUALS, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACE, p.parseStructLiteral)
	p.registerInfix(token.DOT, p.parseMemberExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
}

func (p *Parser) Errors() []string {
	errors := make([]string, len(p.diagnostics))
	for i, diagnostic := range p.diagnostics {
		errors[i] = diagnostic.Error()
	}
	return errors
}

func (p *Parser) Diagnostics() []Diagnostic {
	diagnostics := make([]Diagnostic, len(p.diagnostics))
	for i, diagnostic := range p.diagnostics {
		diagnostics[i] = diagnostic
		diagnostics[i].Expected = append([]token.TokenType(nil), diagnostic.Expected...)
	}
	return diagnostics
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	if t == token.ILLEGAL {
		p.addLexicalDiagnostic(p.curToken)
		return
	}
	msg := fmt.Sprintf("expected expression, found %s", describeToken(p.curToken))
	p.addDiagnostic("P1002", msg, p.curToken)
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf(
		"expected %s, found %s",
		token.GetTokenName(int(t)),
		describeToken(p.peekToken),
	)

	p.addDiagnostic("P1001", msg, p.peekToken, t)
}
