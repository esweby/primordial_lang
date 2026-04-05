package parser

import (
	"fmt"
	"strconv"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/token"
)

type (
	prefixParseFn func() ast.Expression
	infixParseFn func(ast.Expression) ast.Expression
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
)

var systemTypes = map[string]bool{
	"boolean": true,
	"string": true,
	"int": true,
	"int8": true,
	"uint8": true,
	"int32": true,
	"uint32": true,
	"int64": true,
	"uint64": true,
	"float32": true,
	"float64": true,
}

type Parser struct {
	l *lexer.Lexer
	errors []string

	curToken token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l: l,
		errors: []string{},
	}

	p.nextToken()
	p.nextToken()

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT_LITERAL, p.parseIntegerLiteral)
	
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()

		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.PUB, token.MUT, token.CONST:
		return p.parseDeclareStatement() 
	case token.IDENT:
		// peekTokenIs colon is fairly safe to use as other usages of ident: will be 
		// covered within declare statements and not at this initial catch level 
		if p.peekTokenIs(token.COLON) || p.peekTokenIs(token.DECLARE) {
			return p.parseDeclareStatement()
		}

		return p.parseExpressionStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// VARIABLE DECLARATION
func (p *Parser) parseDeclareStatement() *ast.DeclareStatement {
	stmt := &ast.DeclareStatement{
		Public: false,
		Mutable: false,
		Constant: false,
	}

	// Enforces ordering that pub is first
	if p.curTokenIs(token.PUB) {
		stmt.Public = true
		p.nextToken()
	}

	// Enforces only being const or mut
	if p.curTokenIs(token.CONST) {
		stmt.Constant = true
		p.nextToken()
	} else if p.curTokenIs(token.MUT) {
		stmt.Mutable = true
		p.nextToken()
	}

	if !p.curTokenIs(token.IDENT) {
		// Provide error message for improper variable declaration
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	if p.curTokenIs(token.COLON) {
		// Only taking system defined types for the moment
		p.nextToken()
		expectedType := p.curToken.Literal
		if _, ok := systemTypes[expectedType]; !ok {
			// error for invalid type
			return nil
		}

		stmt.Type = expectedType
		p.nextToken()
	}

	if !p.curTokenIs(token.DECLARE) {
		return nil
	}

	stmt.Token = p.curToken
	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)
	// Will need to check that this is needed once full expression
	// parsing is done 
	p.nextToken()

	//
	// Unhappy with this as it should be able to peekToken()
	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// RETURN STATEMENTS
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	for !p.curTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// EXPRESSIONS
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]

	if prefix == nil {
		return nil
	}

	leftExp := prefix()

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil 
	}

	lit.Value = value
	return lit
}

// HELPER FUNCTIONS
func (p *Parser) curTokenIs(tokenType token.TokenType) bool {
	return p.curToken.Type == tokenType
}

func (p *Parser) peekTokenIs(tokenType token.TokenType) bool {
	return p.peekToken.Type == tokenType
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

// ERROR HANDLING
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf(
		"expected next token to be %s, but got %s instead", 
		token.GetTokenName(int(t)), 
		token.GetTokenName(int(p.peekToken.Type)),
	)

	p.errors = append(p.errors, msg)
}