package parser

import (
	"fmt"
	"strconv"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/token"
)

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
	openToken := p.curToken
	p.nextToken()
	exp := p.parseExpression(LOWEST)

	if p.peekTokenIs(token.COMMA) {
		names := []*ast.Identifier{}
		first, ok := exp.(*ast.Identifier)
		if !ok {
			p.errors = append(p.errors, "tuple assignment target must contain only identifiers")
			return nil
		}
		names = append(names, first)

		for p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			name, ok := p.parseIdentifier().(*ast.Identifier)
			if !ok || !p.curTokenIs(token.IDENT) {
				p.errors = append(p.errors, "tuple assignment target must contain only identifiers")
				return nil
			}
			names = append(names, name)
		}

		if !p.expectPeek(token.RPAREN) {
			return nil
		}
		return &ast.TupleTargetExpression{Token: openToken, Names: names}
	}

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
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)
	return expression
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

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
	block := &ast.BlockExpression{Token: p.curToken, Statements: []ast.Statement{}}

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

func (p *Parser) parseMemberExpression(receiver ast.Expression) ast.Expression {
	expression := &ast.MemberExpression{Token: p.curToken, Receiver: receiver}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	expression.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	return expression
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.errors = append(p.errors, fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
		return nil
	}

	literal.Value = value
	return literal
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}
