package parser

import (
	"fmt"

	"github.com/esweby/primordial_lang/ast"
	"github.com/esweby/primordial_lang/token"
)

func (p *Parser) parseMapLiterals() ast.Expression {
	m := &ast.MapLiteral{
		Token: p.curToken,
	}

	mt, ok := p.parseMapType()
	if !ok {
		return nil
	}

	m.Type = mt
	p.nextToken()

	if !p.curTokenIs(token.LBRACE) {
		before := len(p.diagnostics)
		p.ensureDiagnostic(before, "P2001", "expected '{' after type declaration", p.curToken)
		return nil
	}

	pairs := []*ast.MapLiteralPair{}
	p.nextToken()
	if p.curTokenIs(token.RBRACE) {
		m.Pairs = pairs
		return m
	}

	for {
		pair := p.consumeMapPair()
		if pair == nil {
			before := len(p.diagnostics)
			p.ensureDiagnostic(before, "P2002", "expected consumeable pair", p.curToken)
			return nil
		}

		pt, ok := pair.(*ast.MapLiteralPair)
		if !ok {
			before := len(p.diagnostics)
			p.ensureDiagnostic(before, "P2003", "expected consumedable pair to be type *ast.MapLiteralPair", p.curToken)
			return nil
		}

		pairs = append(pairs, pt)
		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			break
		}

		if !p.expectPeek(token.COMMA) {
			return nil
		}

		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			break
		}

		p.nextToken()
	}

	m.Pairs = pairs
	return m
}

func (p *Parser) consumeMapPair() ast.Node {
	mlp := &ast.MapLiteralPair{
		Token: p.curToken,
	}

	key := p.parseExpression(LOWEST)
	if key == nil {
		msg := fmt.Sprintf("expected key expression, found %s", describeToken(p.curToken))
		p.addDiagnostic("P2005", msg, p.curToken)
		return nil
	}
	mlp.Key = key

	p.nextToken()
	if !p.curTokenIs(token.COLON) {
		msg := fmt.Sprintf("expected ':' after key expression, found %s", describeToken(p.curToken))
		p.addDiagnostic("P2006", msg, p.curToken, token.COLON)
		return nil
	}

	p.nextToken()
	value := p.parseExpression(LOWEST)
	if value == nil {
		p.addDiagnostic("P2007", "expected expression after ':'", p.curToken)
		return nil
	}
	mlp.Value = value

	return mlp
}
