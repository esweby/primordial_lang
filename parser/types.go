package parser

import (
	"strconv"

	"github.com/esweby/primordial_lang/token"
	"github.com/esweby/primordial_lang/types"
)

func (p *Parser) parseCurrentType() (types.Type, bool) {
	if !p.curTokenIs(token.IDENT) {
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
		return nil, false
	}

	length, err := strconv.ParseInt(p.curToken.Literal, 10, 0)
	if err != nil {
		return nil, false
	}

	p.nextToken()
	if !p.curTokenIs(token.RBRACKET) {
		return nil, false
	}

	p.nextToken()
	elementType, ok := p.parseCurrentType()
	if !ok {
		return nil, false
	}

	return types.NewArray(elementType, int(length)), true
}
