package parser

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/token"
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
		p.addDiagnostic("P1201", "expected function name, found "+describeToken(p.curToken), p.curToken, token.IDENT)
		return nil
	}

	fn.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.peekTokenIs(token.LPAREN) {
		p.addDiagnostic("P1202", "expected '(' after function name, found "+describeToken(p.peekToken), p.peekToken, token.LPAREN)
		return nil
	}

	p.nextToken()

	var err error
	before := len(p.diagnostics)
	fn.Parameters, err = p.parseFunctionParameters()
	if err != nil {
		p.ensureDiagnostic(before, "P1203", err.Error(), p.curToken)
		return nil
	}

	if p.peekTokenIs(token.COLON) {
		before = len(p.diagnostics)
		fn.ReturnTypes, err = p.parseReturnTypes()
		if err != nil {
			p.ensureDiagnostic(before, "P1204", err.Error(), p.curToken)
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
		p.addDiagnostic("P1202", "expected '(' after 'fn', found "+describeToken(p.peekToken), p.peekToken, token.LPAREN)
		return nil
	}

	p.nextToken()

	var err error
	before := len(p.diagnostics)
	fn.Parameters, err = p.parseFunctionParameters()
	if err != nil {
		p.ensureDiagnostic(before, "P1203", err.Error(), p.curToken)
		return nil
	}

	if p.peekTokenIs(token.COLON) {
		before = len(p.diagnostics)
		fn.ReturnTypes, err = p.parseReturnTypes()
		if err != nil {
			p.ensureDiagnostic(before, "P1204", err.Error(), p.curToken)
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
			p.addDiagnostic("P1203", "expected parameter name, found "+describeToken(p.curToken), p.curToken, token.IDENT)
			return nil, fmt.Errorf("expected parameter name, found %s", describeToken(p.curToken))
		}

		param.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

		p.nextToken()
		paramType, ok := p.parseCurrentType()
		if !ok {
			return nil, fmt.Errorf("expected parameter type for %s", param.Name.Value)
		}

		param.Type = paramType
		params = append(params, param)

		if p.peekTokenIs(token.RPAREN) {
			break
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			continue
		}
		p.addDiagnostic(
			"P1203",
			fmt.Sprintf("expected ',' or ')' after parameter %s, found %s", param.Name.Value, describeToken(p.peekToken)),
			p.peekToken,
			token.COMMA,
			token.RPAREN,
		)
		return nil, fmt.Errorf("expected ',' or ')' after parameter %s, found %s", param.Name.Value, describeToken(p.peekToken))
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
			p.addDiagnostic("P1204", "expected return type, found "+describeToken(p.curToken), p.curToken, token.IDENT)
			return nil, fmt.Errorf("expected return type, found %s", describeToken(p.curToken))
		}

		returnType, ok := p.parseCurrentType()
		if !ok {
			return nil, fmt.Errorf("unknown type %s", p.curToken.Literal)
		}

		returnTypes = append(returnTypes, &ast.ReturnType{Type: returnType})

		if p.peekTokenIs(token.LBRACE) {
			break
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
			p.nextToken()
			continue
		}
		p.addDiagnostic(
			"P1204",
			"expected ',' or '{' after return type, found "+describeToken(p.peekToken),
			p.peekToken,
			token.COMMA,
			token.LBRACE,
		)
		return nil, fmt.Errorf("expected ',' or '{' after return type, found %s", describeToken(p.peekToken))
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
	before := len(p.diagnostics)
	argument := p.parseExpression(LOWEST)
	if argument == nil {
		p.ensureDiagnostic(before, "P1210", "expected function argument", p.curToken)
		return nil
	}
	args = append(args, argument)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		before = len(p.diagnostics)
		argument = p.parseExpression(LOWEST)
		if argument == nil {
			p.ensureDiagnostic(before, "P1210", "expected function argument after ','", p.curToken)
			return nil
		}
		args = append(args, argument)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}
