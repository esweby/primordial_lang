package parser

import (
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
	curToken token.Token
	peekToken token.Token
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}

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
	case token.LTAG:
		if p.isDeclareStatement() {
			return p.parseDeclareStatementWithOptions()
		} else {

		}
	case token.IDENT:
		if p.peekToken.Type == token.DECLARE {
			return p.parseDeclareStatement()
		}
	}

	return nil
}

// VARIABLE DECLARATION
func (p *Parser) isDeclareStatement() bool {
	// Is it a type which are not reserved keywords
	// Still bad form to use them though
	isTypeIdentifier := p.peekTokenIs(token.IDENT) && variableTypes[p.peekToken.Literal]

	if p.peekTokenIs(token.MUT) || p.peekTokenIs(token.PUB) || isTypeIdentifier {
		return true
	} 

	return false
}

func (p *Parser) parseDeclareStatementWithOptions() *ast.DeclareStatement {
	// Enters on a < 
	// Variable can be anything from
	// - <mut, pub, int32> || <pub, mut, int32>
	// - <pub, mut> || <mut, pub>
	// - <pub, int32> || <int32, pub>
	// - <pub> || <mut> || <int32>

	stmt := &ast.DeclareStatement{}

	// As curr token is < then we move it along one to fill the list in 
	p.nextToken()

	optNames := []string{"mut", "pub", "type"}

	opts := map[string]int32{
		"mut": 0,
		"pub": 0,
		"type": 0,
	}

	var varType string

	for {
		// control checks
		if p.currTokenIs(token.RTAG) {
			// exit condition for loop 
			p.nextToken()
			break
		} else if p.currTokenIs(token.COMMA) {
			if p.peekTokenIs(token.RTAG) {
				// TODO: Indicate error for having a comma followed by a rtag <mut,>
				return nil
			}
			p.nextToken() // is delimiter, go next

		// VALID CHECKS
		} else if p.currTokenIs(token.MUT) {
			opts["mut"] += 1
			p.nextToken()
		} else if p.currTokenIs(token.PUB) {
			opts["pub"] += 1
			p.nextToken()
		} else if _, ok := variableTypes[p.curToken.Literal]; ok {
			opts["type"] += 1
			varType = p.curToken.Literal
			p.nextToken()
		// INVALID CHECK
		} else {
			// TODO: Indicate error for unrecognised identifier where mut, pub, or a variable type should be
			return nil
		}
	}

	for _, opt := range optNames {
		if opts[opt] > 1 {
			// TODO: Error message
			return nil
		}
	}
	
	if opts["mut"] == 1 {
		stmt.Mutable = true
	}
	
	if opts["pub"] == 1 {
		stmt.Public = true
	}

	if _, ok := variableTypes[varType]; ok {
		stmt.Type = varType
	}

	if !p.currTokenIs(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.DECLARE) {
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

func (p *Parser) parseDeclareStatement() *ast.DeclareStatement {
	// Entering on an identifier
	stmt := &ast.DeclareStatement{
		Mutable: false,
		Public: false,
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	// This should now be the declare token
	if !p.expectPeek(token.DECLARE) {
		return nil
	}

	stmt.Token = p.curToken

	// This should be expression
	p.nextToken()

	// TODO: Insert expression evaluation

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// HELPER FUNCTIONS

func (p *Parser) currTokenIs(tokenType token.TokenType) bool {
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

	return false
}