package parser

import (
	"strconv"

	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

func (p *Parser) parseType() (types.Type, bool) {
	switch p.curToken.Type {
	case token.MAP:
		return p.parseMapType()
	case token.LBRACKET:
		return p.parseCollectionType()
	default:
		return p.parseCurrentType()
	}
}

func (p *Parser) parseMapType() (types.Type, bool) {
	if !p.curTokenIs(token.MAP) {
		p.addDiagnostic("P1406", "expected map token", p.curToken)
		return nil, false
	}

	p.nextToken()
	if !p.curTokenIs(token.LBRACKET) {
		p.addDiagnostic("P1406", "expected '[' after map", p.curToken)
		return nil, false
	}

	p.nextToken()
	keyType, ok := p.parseCurrentType()
	if !ok {
		before := len(p.diagnostics)
		p.ensureDiagnostic(before, "P1407", "expected type after '['", p.curToken)
		return nil, false
	}

	p.nextToken()
	if !p.curTokenIs(token.RBRACKET) {
		p.addDiagnostic("P1408", "expected ']' after type", p.curToken)
		return nil, false
	}

	p.nextToken()
	mt := &types.Map{
		Key: keyType,
	}

	vt, ok := p.parseType()
	if !ok {
		before := len(p.diagnostics)
		p.ensureDiagnostic(before, "P1407", "expected type after ']'", p.curToken)
		return nil, false
	}
	mt.Value = vt

	return mt, true
}

func (p *Parser) parseCollectionType() (types.Type, bool) {
	p.nextToken()
	if p.curTokenIs(token.RBRACKET) {
		p.nextToken()
		elementType, ok := p.parseCurrentType()
		if !ok {
			return nil, false
		}

		return types.NewSlice(elementType), true
	}

	if !p.curTokenIs(token.INT_LITERAL) {
		p.addDiagnostic("P1403", "expected array length or ']', found "+describeToken(p.curToken), p.curToken, token.INT_LITERAL, token.RBRACKET)
		return nil, false
	}

	length, err := strconv.ParseInt(p.curToken.Literal, 10, 0)
	if err != nil {
		p.addDiagnostic("P1404", "invalid array length "+p.curToken.Literal, p.curToken)
		return nil, false
	}

	p.nextToken()
	if !p.curTokenIs(token.RBRACKET) {
		p.addDiagnostic("P1405", "expected ']' after array length, found "+describeToken(p.curToken), p.curToken, token.RBRACKET)
		return nil, false
	}

	p.nextToken()
	elementType, ok := p.parseCurrentType()
	if !ok {
		return nil, false
	}

	return types.NewArray(elementType, int(length)), true
}

func (p *Parser) parseCurrentType() (types.Type, bool) {
	if !p.curTokenIs(token.IDENT) {
		p.addDiagnostic("P1401", "expected type name, found "+describeToken(p.curToken), p.curToken, token.IDENT)
		return nil, false
	}

	if builtin, ok := types.GetBuiltin(p.curToken.Literal); ok {
		return builtin, true
	}

	return &types.Named{
		CustomName: p.curToken.Literal,
		Underlying: types.InvalidType,
	}, true
}
