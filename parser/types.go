package parser

import (
	"strconv"

	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

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

// parseTypeAfterColon parses a scalar, array, or slice type annotation. It
// expects curToken to be the colon and leaves curToken on the final type token.
func (p *Parser) parseTypeAfterColon() (types.Type, bool) {
	if !p.curTokenIs(token.COLON) {
		p.addDiagnostic("P1402", "expected ':' before type annotation, found "+describeToken(p.curToken), p.curToken, token.COLON)
		return nil, false
	}

	p.nextToken()
	if !p.curTokenIs(token.LBRACKET) {
		return p.parseCurrentType()
	}

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
