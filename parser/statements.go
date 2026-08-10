package parser

import (
	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/token"
)

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.PUB:
		if p.peekTokenIs(token.FN) {
			return p.parseFunctionStatement()
		}
		return p.parseDeclareStatement()
	case token.MUT, token.CONST:
		stmt := p.parseDeclareStatement()
		if stmt == nil {
			return nil
		}
		return stmt
	case token.IDENT:
		if p.peekTokenIs(token.COLON) || p.peekTokenIs(token.DECLARE) {
			stmt := p.parseDeclareStatement()
			if stmt == nil {
				return nil
			}
			return stmt
		}

		return p.parseExpressionStatement()
	case token.RETURN:
		stmt := p.parseReturnStatement()
		if stmt == nil {
			return nil
		}
		return stmt
	case token.FN:
		return p.parseFunctionStatement()
	case token.STRUCT:
		return p.parseStructStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseDeclareStatement() *ast.DeclareStatement {
	stmt := &ast.DeclareStatement{}

	if p.curTokenIs(token.PUB) {
		stmt.Public = true
		p.nextToken()
	}

	if p.curTokenIs(token.CONST) {
		stmt.Constant = true
		p.nextToken()
	} else if p.curTokenIs(token.MUT) {
		stmt.Mutable = true
		p.nextToken()
	}

	if !p.curTokenIs(token.IDENT) {
		p.addDiagnostic("P1101", "expected declaration name, found "+describeToken(p.curToken), p.curToken, token.IDENT)
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	if p.curTokenIs(token.COLON) {
		before := len(p.diagnostics)
		declaredType, ok := p.parseTypeAfterColon()
		if !ok {
			p.ensureDiagnostic(before, "P1401", "expected type after ':'", p.curToken, token.IDENT)
			return nil
		}

		stmt.Type = declaredType
		p.nextToken()
	}

	if !p.curTokenIs(token.DECLARE) {
		p.addDiagnostic("P1102", "expected ':=' after declaration name, found "+describeToken(p.curToken), p.curToken, token.DECLARE)
		return nil
	}

	stmt.Token = p.curToken
	p.nextToken()
	before := len(p.diagnostics)
	stmt.Value = p.parseExpression(LOWEST)
	if stmt.Value == nil {
		p.ensureDiagnostic(before, "P1103", "expected declaration value", p.curToken)
		return nil
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()
	if p.curTokenIs(token.SEMICOLON) {
		stmt.ReturnValues = []ast.Expression{}
		return stmt
	}

	returnValues := []ast.Expression{}
	for {
		before := len(p.diagnostics)
		expr := p.parseExpression(LOWEST)
		if expr == nil {
			p.ensureDiagnostic(before, "P1701", "expected expression after 'return'", p.curToken)
			return nil
		}

		returnValues = append(returnValues, expr)
		if !p.peekTokenIs(token.COMMA) {
			break
		}
		p.nextToken()
		p.nextToken()
	}

	stmt.ReturnValues = returnValues

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
		return stmt
	}
	if !p.curTokenIs(token.SEMICOLON) {
		p.addDiagnostic("P1702", "expected ';' after return values, found "+describeToken(p.peekToken), p.peekToken, token.SEMICOLON)
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	if stmt.Expression == nil {
		return nil
	}

	if target, ok := stmt.Expression.(*ast.TupleTargetExpression); ok &&
		(p.peekTokenIs(token.DECLARE) || p.peekTokenIs(token.ASSIGN)) {
		p.nextToken()
		operator := p.curToken
		p.nextToken()
		before := len(p.diagnostics)
		value := p.parseExpression(LOWEST)
		if value == nil {
			p.ensureDiagnostic(before, "P1603", "expected value after tuple assignment operator", p.curToken)
			return nil
		}

		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}

		if operator.Type == token.DECLARE {
			return &ast.TupleDeclareStatement{Token: operator, Names: target.Names, Value: value}
		}
		return &ast.TupleAssignStatement{Token: operator, Names: target.Names, Value: value}
	}

	if _, ok := stmt.Expression.(*ast.TupleTargetExpression); ok {
		p.addDiagnostic("P1601", "tuple target must be followed by ':=' or '='", p.peekToken, token.DECLARE, token.ASSIGN)
		return nil
	}

	if p.peekTokenIs(token.ASSIGN) {
		switch stmt.Expression.(type) {
		case *ast.Identifier, *ast.MemberExpression:
		default:
			p.addDiagnostic("P1602", "assignment target must be an identifier or member expression", p.curToken)
			return nil
		}

		p.nextToken()
		operator := p.curToken
		p.nextToken()
		before := len(p.diagnostics)
		value := p.parseExpression(LOWEST)
		if value == nil {
			p.ensureDiagnostic(before, "P1603", "expected assignment value", p.curToken)
			return nil
		}

		assignment := &ast.AssignStatement{
			Token:  operator,
			Target: stmt.Expression,
			Value:  value,
		}
		if name, ok := stmt.Expression.(*ast.Identifier); ok {
			assignment.Name = name
		}

		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}

		return assignment
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
