package parser

import (
	"strconv"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
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

		if p.peekTokenIs(token.ASSIGN) {
			return p.parseAssignStatement()
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
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	p.nextToken()

	if p.curTokenIs(token.COLON) {
		p.nextToken()

		if p.curTokenIs(token.LBRACKET) {
			p.nextToken()
			if p.curTokenIs(token.INT_LITERAL) {
				length := p.curToken.Literal
				p.nextToken()
				if !p.curTokenIs(token.RBRACKET) {
					return nil
				}

				p.nextToken()
				builtin, ok := types.GetBuiltin(p.curToken.Literal)
				if !ok {
					return nil
				}
				parsedInt, err := strconv.ParseInt(length, 10, 0)
				if err != nil {
					return nil
				}

				stmt.Type = types.NewArray(builtin, int(parsedInt))
			} else if p.curTokenIs(token.RBRACKET) {
				p.nextToken()
				builtin, ok := types.GetBuiltin(p.curToken.Literal)
				if !ok {
					return nil
				}

				stmt.Type = types.NewSlice(builtin)
			} else {
				return nil
			}
		} else {
			builtin, ok := types.GetBuiltin(p.curToken.Literal)
			if !ok {
				return nil
			}

			stmt.Type = builtin
		}
		p.nextToken()
	}

	if !p.curTokenIs(token.DECLARE) {
		return nil
	}

	stmt.Token = p.curToken
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseAssignStatement() *ast.AssignStatement {
	stmt := &ast.AssignStatement{
		Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	p.nextToken()
	stmt.Token = p.curToken
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}
	p.nextToken()

	returnValues := []ast.Expression{}
	for {
		expr := p.parseExpression(LOWEST)
		if expr == nil {
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

	for !p.curTokenIs(token.SEMICOLON) && !p.curTokenIs(token.EOF) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if target, ok := stmt.Expression.(*ast.TupleTargetExpression); ok &&
		(p.peekTokenIs(token.DECLARE) || p.peekTokenIs(token.ASSIGN)) {
		p.nextToken()
		operator := p.curToken
		p.nextToken()
		value := p.parseExpression(LOWEST)

		if p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}

		if operator.Type == token.DECLARE {
			return &ast.TupleDeclareStatement{Token: operator, Names: target.Names, Value: value}
		}
		return &ast.TupleAssignStatement{Token: operator, Names: target.Names, Value: value}
	}

	if _, ok := stmt.Expression.(*ast.TupleTargetExpression); ok {
		p.errors = append(p.errors, "tuple target must be followed by ':=' or '='")
		return nil
	}

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
