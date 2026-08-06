package parser

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

func (p *Parser) parseFunctionStatement() ast.Statement {
	fn := &ast.FunctionStatement{}

	if p.curTokenIs(token.PUB) {
		fn.Public = true
		p.nextToken()
	}

	fn.Token = p.curToken
	p.nextToken()

	if !p.curTokenIs(token.IDENT) {
		return nil
	}

	fn.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.peekTokenIs(token.LPAREN) {
		return nil
	}

	p.nextToken()

	var err error
	fn.Parameters, err = p.parseFunctionParameters()
	if err != nil {
		return nil
	}

	if p.peekTokenIs(token.COLON) {
		fn.ReturnTypes, err = p.parseReturnTypes()
		if err != nil {
			return nil
		}
	} else {
		fn.ReturnTypes = []*ast.ReturnType{}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fn.Body = p.parseBlockExpression()
	return fn
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.curToken}

	if !p.peekTokenIs(token.LPAREN) {
		return nil
	}

	p.nextToken()

	var err error
	fn.Parameters, err = p.parseFunctionParameters()
	if err != nil {
		return nil
	}

	if p.peekTokenIs(token.COLON) {
		fn.ReturnTypes, err = p.parseReturnTypes()
		if err != nil {
			return nil
		}
	} else {
		fn.ReturnTypes = []*ast.ReturnType{}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	fn.Body = p.parseBlockExpression()
	return fn
}

func (p *Parser) parseFunctionParameters() ([]*ast.Parameter, error) {
	params := []*ast.Parameter{}

	p.nextToken()
	if p.curTokenIs(token.RPAREN) {
		return params, nil
	}

	for {
		param := &ast.Parameter{}

		if !p.curTokenIs(token.IDENT) {
			return nil, fmt.Errorf("expected parameter name. got=%v", p.peekToken.Type)
		}

		param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		p.nextToken()
		builtin, ok := types.GetBuiltin(p.curToken.Literal)
		if !ok {
			return nil, fmt.Errorf("unknown type %s", p.curToken.Literal)
		}

		param.Type = builtin
		params = append(params, param)

		if p.peekTokenIs(token.RPAREN) {
			break
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			continue
		}
		return nil, fmt.Errorf("expected ',' or ')', got %v", p.peekToken.Type)
	}

	p.nextToken()
	return params, nil
}

func (p *Parser) parseReturnTypes() ([]*ast.ReturnType, error) {
	if !p.expectPeek(token.COLON) {
		return nil, fmt.Errorf("expected ':' for return types")
	}
	p.nextToken()

	returnTypes := []*ast.ReturnType{}
	for {
		if !p.curTokenIs(token.IDENT) {
			return nil, fmt.Errorf("expected identifier return type, got %v", p.curToken.Type)
		}

		builtin, ok := types.GetBuiltin(p.curToken.Literal)
		if !ok {
			return nil, fmt.Errorf("unknown type %s", p.curToken.Literal)
		}

		returnTypes = append(returnTypes, &ast.ReturnType{Type: builtin})

		if p.peekTokenIs(token.LBRACE) {
			break
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			continue
		}
		return nil, fmt.Errorf("expected ',' or '{' after return type, got %v", p.peekToken.Type)
	}

	return returnTypes, nil
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
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
