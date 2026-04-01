package parser

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/token"
)

var variableTypes = map[string]bool{
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
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l: l,
		errors: []string{},
	}

	p.nextToken()
	p.nextToken()
	
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
	case token.RETURN:
		return p.parseReturnStatement()
	}

	return nil
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
		if _, ok := variableTypes[expectedType]; !ok {
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

	// TODO: Insert expression evaluation

	if p.peekTokenIs(token.SEMICOLON) {
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