package parser

import (
	"fmt"
	"strconv"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/lexer"
	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
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
)

var precedences = map[token.TokenType]int{
	token.EQUALS:        EQUALS,
	token.NOT_EQUALS:    EQUALS,
	token.LTAG:          LESSGREATER,
	token.RTAG:          LESSGREATER,
	token.PLUS:          SUM,
	token.MINUS:         SUM,
	token.FORWARD_SLASH: PRODUCT,
	token.ASTERIK:       PRODUCT,
	token.LPAREN:        CALL,
}

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.registerPrefixFns()
	p.registerInfixFns()

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
	case token.PUB:
		if p.peekTokenIs(token.FN) {
			return p.parseFunctionStatement()
		}
		return p.parseDeclareStatement()
	case token.MUT, token.CONST:
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
	case token.FN:
		return p.parseFunctionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", token.GetTokenName(int(t)))
	p.errors = append(p.errors, msg)
}

// VARIABLE DECLARATION
func (p *Parser) parseDeclareStatement() *ast.DeclareStatement {
	stmt := &ast.DeclareStatement{
		Public:   false,
		Mutable:  false,
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
		p.nextToken()
		expectedType := p.curToken.Literal
		builtin, ok := types.GetBuiltin(expectedType)
		if !ok {
			// error for invalid type
			return nil
		}

		stmt.Type = builtin
		p.nextToken()
	}

	if !p.curTokenIs(token.DECLARE) {
		return nil
	}

	stmt.Token = p.curToken
	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	for !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.EOF) && !p.curTokenIs(token.RBRACE) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseFunctionStatement() ast.Statement {
	fn := &ast.FunctionStatement{}

	if p.curTokenIs(token.PUB) {
		fn.Public = true
		p.nextToken()
	}

	fn.Token = p.curToken
	p.nextToken()

	if !p.curTokenIs(token.IDENT) {
		// TODO: Error processing
		return nil
	}

	fn.Name = &ast.Identifier{
		Token: p.curToken,
		Value: p.curToken.Literal,
	}

	if !p.peekTokenIs(token.LPAREN) {
		return nil
	}

	p.nextToken() // (

	var err error
	fn.Parameters, err = p.parseFunctionParameters()
	if err != nil {
		// TODO: Handle error
		return nil
	}

	// is )
	if p.peekTokenIs(token.COLON) {
		fn.ReturnTypes, err = p.parseReturnTypes()
		if err != nil {
			// TODO: Handle errors
			return nil
		}
	} else {
		fn.ReturnTypes = []*ast.ReturnType{}
	}

	// should be a type ident
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fn.Body = p.parseBlockExpression()

	return fn
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.curToken}

	// is fn
	if !p.peekTokenIs(token.LPAREN) {
		return nil
	}

	// is (
	p.nextToken()

	var err error
	fn.Parameters, err = p.parseFunctionParameters()
	if err != nil {
		// TODO: Handle error
		return nil
	}

	// is )
	if p.peekTokenIs(token.COLON) {
		fn.ReturnTypes, err = p.parseReturnTypes()
		if err != nil {
			// TODO: Handle errors
			return nil
		}
	} else {
		fn.ReturnTypes = []*ast.ReturnType{}
	}

	// should be a type ident
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fn.Body = p.parseBlockExpression()

	return fn
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseFunctionParameters() ([]*ast.Parameter, error) {
	params := []*ast.Parameter{}

	// currently ( looking for ident or )
	p.nextToken()
	if p.curTokenIs(token.RPAREN) {
		// empty list returning on )
		return params, nil
	}

	for {
		param := &ast.Parameter{}

		if !p.curTokenIs(token.IDENT) {
			// if its an error we're currently throwing a full parsing error
			return nil, fmt.Errorf("expected parameter name. got=%v", p.peekToken.Type)
		}

		param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		// This will be the type
		p.nextToken()
		expectedType := p.curToken.Literal
		builtin, ok := types.GetBuiltin(expectedType)
		if !ok {
			// error for invalid type
			return nil, fmt.Errorf("unknown type %s", expectedType)
		}

		param.Type = builtin


		params = append(params, param)

		// After a parameter, we expect either ',' or ')'
		if p.peekTokenIs(token.RPAREN) {
			// Current ident so will go to ) on the outer loop
			break
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken() // consume ','
			// want ident so go for that
			p.nextToken()
			continue
		}
		return nil, fmt.Errorf("expected ',' or ')', got %v", p.peekToken.Type)
	}

	p.nextToken() // move to rparen

	return params, nil
}

func (p *Parser) parseReturnTypes() ([]*ast.ReturnType, error) {
    // Expect and consume the colon
    if !p.expectPeek(token.COLON) {
        return nil, fmt.Errorf("expected ':' for return types")
    }
    // Move to the first return type identifier
    p.nextToken()

    returnTypes := []*ast.ReturnType{}

    for {
        if !p.curTokenIs(token.IDENT) {
            return nil, fmt.Errorf("expected identifier return type, got %v", p.curToken.Type)
        }

        // Read the current identifier as the type
        typeLiteral := p.curToken.Literal
        builtin, ok := types.GetBuiltin(typeLiteral)
        if !ok {
            return nil, fmt.Errorf("unknown type %s", typeLiteral)
        }

        rt := &ast.ReturnType{Type: builtin}
        returnTypes = append(returnTypes, rt)

        // Check what comes after this type
        if p.peekTokenIs(token.LBRACE) {
            // Done – next is the function body's '{'
            break
        }
        if p.peekTokenIs(token.COMMA) {
            p.nextToken() // consume ','
            p.nextToken() // move to next type identifier
            continue
        }
        return nil, fmt.Errorf("expected ',' or '{' after return type, got %v", p.peekToken.Type)
    }

    return returnTypes, nil
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()

		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{
		Token: p.curToken,
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()

	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Body = p.parseBlockExpression()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()
		p.nextToken()

		if p.curTokenIs(token.LBRACE) {
			expression.Else = p.parseBlockExpression()
		} else if p.curTokenIs(token.IF) {
			expression.Else = p.parseIfExpression()
		} else {
			return nil
		}
	}

	return expression
}

func (p *Parser) parseBlockExpression() *ast.BlockExpression {
	block := &ast.BlockExpression{
		Token: p.curToken,
	}

	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		p.nextToken()
	}

	return block
}

// RETURN STATEMENTS
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	
	stmt.ReturnValue = p.parseExpression(LOWEST)

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
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]

		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	exp := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)

	return exp
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

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// HELPER FUNCTIONS
func (p *Parser) curTokenIs(tokenType token.TokenType) bool {
	return p.curToken.Type == tokenType
}

func (p *Parser) peekTokenIs(tokenType token.TokenType) bool {
	return p.peekToken.Type == tokenType
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
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
